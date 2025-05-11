package application

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/arniknz/calculator_go_5/pkg/calculator"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	Port                string
	TimeAddition        int
	TimeSubtraction     int
	TimeMultiplications int
	TimeDivisions       int
	Debug               bool
}

func Configure() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	a, _ := strconv.Atoi(os.Getenv("TIME_ADDITION_MS"))
	if a == 0 {
		a = 20
	}
	s, _ := strconv.Atoi(os.Getenv("TIME_SUBTRACTION_MS"))
	if s == 0 {
		s = 20
	}
	m, _ := strconv.Atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"))
	if m == 0 {
		m = 70
	}
	d, _ := strconv.Atoi(os.Getenv("TIME_DIVISIONS_MS"))
	if d == 0 {
		d = 70
	}
	debug := false
	debugStr := os.Getenv("DEBUG")
	if debugStr != "" {
		debug, _ = strconv.ParseBool(debugStr)
	}

	return &Config{
		Port:                port,
		TimeAddition:        a,
		TimeSubtraction:     s,
		TimeMultiplications: m,
		TimeDivisions:       d,
		Debug:               debug,
	}
}

type Orchestrator struct {
	Config    *Config
	exprStore map[string]*Expression
	taskStore map[string]*Task
	taskQueue []*Task
	m         sync.Mutex
	database  *sqlx.DB
	taskCnt   int64
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		Config:    Configure(),
		exprStore: make(map[string]*Expression),
		taskStore: make(map[string]*Task),
		taskQueue: make([]*Task, 0),
		database:  NewDB(),
	}
}

type Expression struct {
	ID     string   `json:"id"`
	Expr   string   `json:"expression"`
	Status string   `json:"status"`
	Result *float64 `json:"result,omitempty"`
	AST    *ASTNode `json:"-"`
}

type Task struct {
	ID            string   `json:"id"`
	ExprID        string   `json:"-"`
	Arg1          float64  `json:"arg1"`
	Arg2          float64  `json:"arg2"`
	Operation     string   `json:"operation"`
	OperationTime int      `json:"operation_time"`
	Node          *ASTNode `json:"-"`
}

func updateStatusAndResult(db *sql.DB, exprID int, status string, result float64) error {
	_, err := db.Exec("UPDATE expressions SET status=?, result=? WHERE id=?", status, result, exprID)
	if err != nil {
		return err
	}
	return nil
}

func (o *Orchestrator) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var request struct{ Login, Password string }
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "error: Bad request", http.StatusBadRequest)
		return
	}
	hash, err := HashPassword(request.Password)
	if err != nil {
		http.Error(w, "error: Internal server error", http.StatusInternalServerError)
		return
	}
	if _, err := o.database.Exec("INSERT INTO users(login,password) VALUES(?,?)", request.Login, hash); err != nil {
		http.Error(w, "error: User exists", http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (o *Orchestrator) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var request struct{ Login, Password string }

	json.NewDecoder(r.Body).Decode(&request)
	var id int
	var hash string
	err := o.database.Get(&hash, "SELECT password FROM users WHERE login=?", request.Login)
	if err != nil {
		http.Error(w, "error: Invalid creds", http.StatusUnauthorized)
		return
	}
	err = CheckPassword(hash, request.Password)
	if err != nil {
		http.Error(w, "error: Invalid creds", http.StatusUnauthorized)
		return
	}

	o.database.Get(&id, "SELECT id FROM users WHERE login=?", request.Login)
	token, err := CreateToken(id)
	if err != nil {
		http.Error(w, "error: Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (o *Orchestrator) CalculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, calculator.ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Expression string `json:"expression"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Expression == "" {
		http.Error(w, calculator.ErrInvalidBody.Error(), http.StatusUnprocessableEntity)
		return
	}
	ast, err := ParseAST(req.Expression)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error" : "%s"}`, err.Error()), http.StatusUnprocessableEntity)
		return
	}
	user_id := 0
	if o.Config.Debug {
		user_id = 1
	} else {
		user_id = r.Context().Value(user_id_key{}).(int)
	}
	o.m.Lock()
	expr := &Expression{
		Expr:   req.Expression,
		Status: "pending",
		AST:    ast,
	}
	inserted, err := o.database.Exec("INSERT INTO expressions(user_id,expression,status) VALUES(?,?,?)", user_id, expr.Expr, expr.Status)
	if err != nil {
		log.Fatal(err)
	}
	exprID, err := inserted.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	expr.ID = strconv.Itoa(int(exprID))
	o.exprStore[expr.ID] = expr
	o.scheduleTasks(expr)
	o.m.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": expr.ID})
}

func (o *Orchestrator) expressionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, calculator.ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}
	o.m.Lock()
	defer o.m.Unlock()
	for _, expr := range o.exprStore {
		if expr.AST != nil && expr.AST.IsLeaf {
			expr.Status = "completed"
			expr.Result = &expr.AST.Value
			expr_id, err := strconv.Atoi(expr.ID)
			if err != nil {
				log.Fatal(err)
			}
			err = updateStatusAndResult(o.database.DB, expr_id, expr.Status, float64(*expr.Result))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	user_id := 0
	if o.Config.Debug {
		user_id = 1
	} else {
		user_id = r.Context().Value(user_id_key{}).(int)
	}
	var exprs_struct []struct {
		ID     int      `db:"id" json:"id"`
		Expr   string   `db:"expression" json:"expression"`
		Status string   `db:"status" json:"status"`
		Result *float64 `db:"result" json:"result,omitempty"`
	}
	if err := o.database.Select(&exprs_struct, "SELECT id,expression,status,result FROM expressions WHERE user_id=?", user_id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"expressions": exprs_struct})
}

func (o *Orchestrator) expressionByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, calculator.ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Path[len("/api/v1/expressions/"):]
	expr_id, err := strconv.Atoi(id)
	if err != nil {
		log.Fatal(err)
	}
	o.m.Lock()
	expr, ok := o.exprStore[id]
	o.m.Unlock()
	if !ok {
		http.Error(w, calculator.ErrExpressionNotFound.Error(), http.StatusNotFound)
		return
	}
	user_id := 0
	if o.Config.Debug {
		user_id = 1
	} else {
		user_id = r.Context().Value(user_id_key{}).(int)
	}
	if expr.AST != nil && expr.AST.IsLeaf {
		expr.Status = "completed"
		expr.Result = &expr.AST.Value
		err = updateStatusAndResult(o.database.DB, expr_id, expr.Status, float64(*expr.Result))
		if err != nil {
			log.Fatal(err)
		}
	}
	var exprs_struct struct {
		ID     int      `db:"id" json:"id"`
		Expr   string   `db:"expression" json:"expression"`
		Status string   `db:"status" json:"status"`
		Result *float64 `db:"result" json:"result,omitempty"`
	}
	if err := o.database.Get(&exprs_struct, "SELECT id,expression,status,result FROM expressions WHERE user_id=? AND id=?", user_id, expr_id); err != nil {
		http.Error(w, calculator.ErrExpressionNotFound.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"expression": exprs_struct})
}

func (o *Orchestrator) getTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, calculator.ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}
	o.m.Lock()
	defer o.m.Unlock()
	if len(o.taskQueue) == 0 {
		http.Error(w, calculator.ErrTaskNotFound.Error(), http.StatusNotFound)
		return
	}
	task := o.taskQueue[0]
	o.taskQueue = o.taskQueue[1:]
	if expr, exists := o.exprStore[task.ExprID]; exists {
		expr.Status = "progressing"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"task": task})
}

func (o *Orchestrator) postTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, calculator.ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID     string  `json:"id"`
		Result float64 `json:"result"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.ID == "" {
		http.Error(w, calculator.ErrInvalidBody.Error(), http.StatusUnprocessableEntity)
		return
	}
	o.m.Lock()
	task, ok := o.taskStore[req.ID]
	if !ok {
		o.m.Unlock()
		http.Error(w, calculator.ErrTaskNotFound.Error(), http.StatusNotFound)
		return
	}
	task.Node.IsLeaf = true
	task.Node.Value = req.Result
	delete(o.taskStore, req.ID)
	if expr, exists := o.exprStore[task.ExprID]; exists {
		o.scheduleTasks(expr)
		if expr.AST.IsLeaf {
			expr.Status = "completed"
			expr.Result = &expr.AST.Value
		}
	}
	o.m.Unlock()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status" : "Result Accepted"}`))
}

func (o *Orchestrator) scheduleTasks(expr *Expression) {
	var traverse func(node *ASTNode)
	traverse = func(node *ASTNode) {
		if node == nil || node.IsLeaf {
			return
		}
		traverse(node.Left)
		traverse(node.Right)
		if node.Left != nil && node.Right != nil && node.Left.IsLeaf && node.Right.IsLeaf {
			if !node.TaskScheduled {
				o.taskCnt += 1
				taskID := fmt.Sprintf("%d", o.taskCnt)
				var opTime int
				switch node.Operator {
				case "+":
					opTime = o.Config.TimeAddition
				case "-":
					opTime = o.Config.TimeSubtraction
				case "*":
					opTime = o.Config.TimeMultiplications
				case "/":
					opTime = o.Config.TimeDivisions
				default:
					opTime = 100
				}
				task := &Task{
					ID:            taskID,
					ExprID:        expr.ID,
					Arg1:          node.Left.Value,
					Arg2:          node.Right.Value,
					Operation:     node.Operator,
					OperationTime: opTime,
					Node:          node,
				}
				node.TaskScheduled = true
				o.taskStore[taskID] = task
				o.taskQueue = append(o.taskQueue, task)
			}
		}
	}
	traverse(expr.AST)
}

func (o *Orchestrator) StartServer() error {
	m := http.NewServeMux()

	m.HandleFunc("/api/v1/register", o.RegisterHandler)
	m.HandleFunc("/api/v1/login", o.LoginHandler)
	m.HandleFunc("/api/v1/calculate", func(w http.ResponseWriter, r *http.Request) {
		o.AuthMiddleware(http.HandlerFunc(o.CalculateHandler)).ServeHTTP(w, r)
	})
	m.HandleFunc("/api/v1/expressions", func(w http.ResponseWriter, r *http.Request) {
		o.AuthMiddleware(http.HandlerFunc(o.expressionsHandler)).ServeHTTP(w, r)
	})
	m.HandleFunc("/api/v1/expressions/", func(w http.ResponseWriter, r *http.Request) {
		o.AuthMiddleware(http.HandlerFunc(o.expressionByIDHandler)).ServeHTTP(w, r)
	})
	m.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			o.getTaskHandler(w, r)
		} else if r.Method == http.MethodPost {
			o.postTaskHandler(w, r)
		} else {
			http.Error(w, calculator.ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		}
	})
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, calculator.ErrNotFound.Error(), http.StatusNotFound)
	})
	go func() {
		for {
			time.Sleep(2 * time.Second)
			o.m.Lock()
			if len(o.taskQueue) > 0 {
				log.Printf("Pending tasks in queue: %d", len(o.taskQueue))
			}
			o.m.Unlock()
		}
	}()
	return http.ListenAndServe(":"+o.Config.Port, m)
}
