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
		X:  left,
		Y:  right,
		Op: token.ADD,
	}
	leftStep := &evaluableStep{Evaluable: left}
	rightStep := &evaluableStep{Evaluable: right}
	leftStep.SetNext(rightStep)
	binExprStep := &evaluableStep{Evaluable: expr}
	rightStep.SetNext(binExprStep)

	vm := newVM(newEnvironment(nil))
	var here Step = leftStep
	for here != nil {
		here.Take(vm)
		here = here.Next()
	}
	if got, want := vm.frameStack.top().pop().Interface(), "Hello, World!"; got != want {
		t.Errorf("got [%[1]v:%[1]T] want [%[2]v:%[2]T]", got, want)
	}
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
