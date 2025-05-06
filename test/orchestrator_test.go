package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	application "github.com/arniknz/calculator_go_5/internal/app/orchestrator"
)

func TestCalculateHandler(t *testing.T) {
	o := application.NewOrchestrator()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		o.CalculateHandler(w, r)
	})

	reqBody := `{"expression": "1 + 1 + 1 + 1 * 2 * 2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	nr := httptest.NewRecorder()

	handler.ServeHTTP(nr, req)

	if nr.Code != http.StatusCreated {
		t.Errorf("Expected %d, got %d", http.StatusCreated, nr.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(nr.Body).Decode(&resp); err != nil {
		t.Fatalf("Error while decoding %v", err)
	}
	if id, ok := resp["id"]; !ok || id == "" {
		t.Errorf("Expected valid id, got: %v", resp)
	}
}
