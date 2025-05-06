package calculator

import "errors"

var (
	ErrInvalidExpression  = errors.New(`{"error" : "Invalid expression"}`)
	ErrDivisionByZero     = errors.New(`{"error" : "Division by zero"}`)
	ErrMethodNotAllowed   = errors.New(`{"error" : "Method Not Allowed"}`)
	ErrExpressionNotFound = errors.New(`{"error" : "Expression Not Found"}`)
	ErrNotFound           = errors.New(`{"error" : "Not Found"}`)
	ErrTaskNotFound       = errors.New(`{"error" : "Task not found"}`)
	ErrInvalidBody        = errors.New(`{"error" : "Invalid Body"}`)
)
