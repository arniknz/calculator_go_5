package calculator

import (
	"fmt"
	"strconv"
	"strings"
)

func Calc(operation string, a, b float64) (float64, error) {
	expression := fmt.Sprintf("%f %s %f", a, operation, b)
	tokens := tokenize(expression)
	postfix, err := infixToPostfix(tokens)
	if err != nil {
		return 0, err
	}
	return evaluatePostfix(postfix)
}

func tokenize(expr string) []string {
	var tokens []string
	var currentToken strings.Builder

	for _, char := range expr {
		switch char {
		case ' ':
			continue
		case '+', '-', '*', '/', '(', ')':
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, string(char))
		default:
			currentToken.WriteRune(char)
		}
	}

	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}

	return tokens
}

func infixToPostfix(tokens []string) ([]string, error) {
	var output []string
	var operators []string

	for _, token := range tokens {
		if isNumber(token) {
			output = append(output, token)
		} else if token == "(" {
			operators = append(operators, token)
		} else if token == ")" {
			for len(operators) > 0 && operators[len(operators)-1] != "(" {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			if len(operators) == 0 {
				return nil, ErrInvalidExpression
			}
			operators = operators[:len(operators)-1]
		} else if isOperator(token) {
			for len(operators) > 0 && priority(operators[len(operators)-1]) >= priority(token) {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			operators = append(operators, token)
		} else {
			return nil, ErrInvalidExpression
		}
	}

	for len(operators) > 0 {
		if operators[len(operators)-1] == "(" {
			return nil, ErrInvalidExpression
		}
		output = append(output, operators[len(operators)-1])
		operators = operators[:len(operators)-1]
	}

	return output, nil
}

func evaluatePostfix(postfix []string) (float64, error) {
	var stack []float64

	for _, token := range postfix {
		if isNumber(token) {
			num, _ := strconv.ParseFloat(token, 64)
			stack = append(stack, num)
		} else if isOperator(token) {
			if len(stack) < 2 {
				return 0, ErrInvalidExpression
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			switch token {
			case "+":
				stack = append(stack, a+b)
			case "-":
				stack = append(stack, a-b)
			case "*":
				stack = append(stack, a*b)
			case "/":
				if b == 0 {
					return 0, ErrDivisionByZero
				}
				stack = append(stack, a/b)
			default:
				return 0, ErrInvalidExpression
			}
		} else {
			return 0, ErrInvalidExpression
		}
	}

	if len(stack) != 1 {
		return 0, ErrInvalidExpression
	}

	return stack[0], nil
}

func isNumber(token string) bool {
	if _, err := strconv.ParseFloat(token, 64); err == nil {
		return true
	}
	return false
}

func isOperator(token string) bool {
	return token == "+" || token == "-" || token == "*" || token == "/"
}

func priority(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	default:
		return 0
	}
}
