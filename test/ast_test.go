package test

import (
	"fmt"
	"testing"

	application "github.com/arniknz/calculator_go_5/internal/app/orchestrator"
	"github.com/arniknz/calculator_go_5/pkg/calculator"
)

func evalAST(node *application.ASTNode) (float64, error) {
	if node.IsLeaf {
		return node.Value, nil
	}
	left, err := evalAST(node.Left)
	if err != nil {
		return 0, err
	}
	right, err := evalAST(node.Right)
	if err != nil {
		return 0, err
	}
	switch node.Operator {
	case "+":
		return left + right, nil
	case "-":
		return left - right, nil
	case "*":
		return left * right, nil
	case "/":
		if right == 0 {
			return 0, calculator.ErrDivisionByZero
		}
		return left / right, nil
	default:
		return 0, fmt.Errorf("unknown operator: %s", node.Operator)
	}
}

func TestParseASTValid(t *testing.T) {
	tests := []struct {
		expr     string
		expected float64
	}{
		{"1 + 2 - 1", 2},
		{"10000000000000", 10000000000000},
		{"8 + 8 * (9 - 1)", 72},
		{"( 4 / 2 ) + 0", 2},
		{"2 + 3 * 5", 17},
	}
	for _, tc := range tests {
		ast, err := application.ParseAST(tc.expr)
		if err != nil {
			t.Errorf("Unexpected error for expression %s: %v", tc.expr, err)
			continue
		}
		result, err := evalAST(ast)
		if err != nil {
			t.Errorf("AST evaluation error %s: %v", tc.expr, err)
			continue
		}
		if result != tc.expected {
			t.Errorf("Expected %f for expression %s, but got %f", tc.expected, tc.expr, result)
		}
	}
}

func TestParseASTInvalid(t *testing.T) {
	invalidExprs := []string{
		"",
		"1+",
		"(1+2",
		"1++2",
		"abc",
	}
	for _, expr := range invalidExprs {
		_, err := application.ParseAST(expr)
		if err == nil {
			t.Errorf("Expected an error for invalid expression %q, but no error occurred", expr)
		}
	}
}
