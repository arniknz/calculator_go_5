package test

import (
	"testing"

	"github.com/arniknz/calculator_go_5/pkg/calculator"
)

func TestCalc(t *testing.T) {
	testCasesSuccess := []struct {
		name        string
		op          string
		a, b        float64
		expected    float64
		ExpectedErr bool
	}{
		{"test 1", "+", 7, 3, 10, false},
		{"test 2", "-", 6, 6, 0, false},
		{"test 3", "*", 5, 5, 25, false},
		{"test 4", "/", 10, 2, 5, false},
		{"test 5", "+", 10, 0, 10, false},
	}

	for _, testCase := range testCasesSuccess {
		t.Run(testCase.name, func(t *testing.T) {
			val, err := calculator.Calc(testCase.op, testCase.a, testCase.b)
			if err != nil {
				t.Fatalf("successful case %f returns error", val)
			}
			if testCase.ExpectedErr && err == nil {
				t.Errorf("Expected error for operation %s", testCase.op)
			}
			if !testCase.ExpectedErr && val != testCase.expected {
				t.Errorf("Calc(%s, %f, %f) = %f; Expected %f", testCase.op, testCase.a, testCase.b, val, testCase.expected)
			}
		})
	}
}
