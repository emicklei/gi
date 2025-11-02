package internal

import (
	"go/ast"
	"go/token"
	"testing"
)

func TestStepByStep(t *testing.T) {
	left := BasicLit{BasicLit: &ast.BasicLit{Kind: token.STRING, Value: "Hello, "}}
	right := BasicLit{BasicLit: &ast.BasicLit{Kind: token.STRING, Value: "World!"}}
	expr := BinaryExpr{
		X:          left,
		Y:          right,
		BinaryExpr: &ast.BinaryExpr{Op: token.ADD},
	}
	leftStep := &evaluableStep{Evaluable: left}
	rightStep := &evaluableStep{Evaluable: right}
	leftStep.SetNext(rightStep)
	binExprStep := &evaluableStep{Evaluable: expr}
	rightStep.SetNext(binExprStep)

	vm := newVM(newEnvironment(nil))
	vm.isStepping = true
	var here Step = leftStep
	for here != nil {
		t.Log(here)
		here.Eval(vm)
		here = here.Next()
	}
	t.Log("result:", vm.frameStack.top().pop().Interface())
}

func TestEvaluableStep(t *testing.T) {
	lit := BasicLit{BasicLit: &ast.BasicLit{Kind: token.INT, Value: "42"}}
	step1 := &evaluableStep{Evaluable: lit}
	step2 := &evaluableStep{Evaluable: lit}
	set := func(s, n Step) {
		s.SetNext(n)
	}
	set(step1, step2)
	if step1.Next() != step2 {
		t.Errorf("expected step1.Next() to be step2")
	}
}
