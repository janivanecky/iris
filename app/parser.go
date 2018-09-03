package app

import (
	"math"
	"strings"
	"strconv"
	"text/scanner"
)

func precedence(op rune) int {
	switch op {
	case '*', '/', '^':
		return 2
	case '+', '-':
		return 1
	}
	return 0
}

type lexer struct {
	s *scanner.Scanner
	token rune
}

func (l *lexer) next() {
	l.token = l.s.Scan()
}

func (l *lexer) text() string {
	return l.s.TokenText()
}

type Expr interface {
	Eval(env map[string]float64) float64
}

type number float64

func (n number) Eval(env map[string]float64) float64 {
	return float64(n)
}

type binary struct {
	x, y Expr
	op rune
}

func (b binary) Eval(env map[string]float64) float64 {
	switch b.op {
	case '+':
		return b.x.Eval(env) + b.y.Eval(env)
	case '-':
		return b.x.Eval(env) - b.y.Eval(env)
	case '*':
		return b.x.Eval(env) * b.y.Eval(env)
	case '/':
		return b.x.Eval(env) / b.y.Eval(env)
	case '^':
		return math.Pow(b.x.Eval(env), b.y.Eval(env))
	}
	return 0.0
}

type variable struct {
	name string
}

var constants = map[string] float64 {
	"e": math.E,
	"pi": math.Pi,
}

func (v variable) Eval(env map[string]float64) float64 {
	constantValue, ok := constants[v.name]
	if !ok {
		return env[v.name]
	}
	return constantValue
}

type function struct {
	name string
	arg Expr
}

func (f function) Eval(env map[string]float64) float64 {
	return functions[f.name](f.arg.Eval(env))
}

var functions = map[string]func(float64) float64 {
	"sin": math.Sin,
	"cos": math.Cos,
	"tan": math.Tan,
	"tanh": math.Tanh,
	"asin": math.Asin,
	"acos": math.Acos,
	"abs": math.Abs,
	"exp": math.Exp,
}

func parseSingle(l *lexer) Expr {
	switch l.token {
	case '-':
		l.next()
		value := parseSingle(l)
		if value == nil {
			return nil
		}
		return binary{number(0), value, '-'}
	case scanner.Float, scanner.Int:
		n, err := strconv.ParseFloat(l.text(), 64)
		if err != nil {
			return nil
		}
		l.next()
		return number(n)
	case '(':
		l.next()
		exp := parseBinary(l, 1)
		if l.token != ')' {
			return nil
		}
		l.next()
		return exp
	case scanner.Ident:
		isFunction := false
		for functionName, _ := range functions {
			if functionName == l.text() {
				isFunction = true
				break
			}
		}
		var exp Expr
		if isFunction {
			functionName := l.text()
			l.next()
			arg := parseBinary(l, 3)
			if arg == nil {
				return nil
			}
			exp = function{functionName, arg}
		} else {
			exp = variable{l.text()}
			l.next()
		}
		return exp
	}
	return nil
}

func parseUnary(l *lexer) Expr {
	return parseSingle(l)
}

func parseBinary(l *lexer, minPrecedence int) Expr {
	lhs := parseUnary(l)
	if lhs == nil {
		return nil
	}
	for opPrec := precedence(l.token); opPrec >= minPrecedence; {
		op := l.token
		l.next()
		rhs := parseBinary(l, opPrec + 1)
		if rhs == nil {
			return nil
		}
		lhs = binary{lhs, rhs, op}
		opPrec = precedence(l.token)
	}
	return lhs
}

func Parse(text string) Expr {
	s := new(scanner.Scanner)
	s.Init(strings.NewReader(text))
	l := lexer{s, 0}
	
	l.next()
	exp := parseBinary(&l, 1)
	return exp
}