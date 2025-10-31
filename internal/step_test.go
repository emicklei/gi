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
	leftStep := &step{Evaluable: left}
	rightStep := &step{Evaluable: right}
	leftStep.SetNext(rightStep)
	binExprStep := &step{Evaluable: expr}
	rightStep.SetNext(binExprStep)

	vm := newVM(newEnvironment(nil))
	var here Step = leftStep
	for here != nil {
		t.Log(here)
		here.Eval(vm)
		here = here.Next()
	}
	t.Log("result:", vm.frameStack.top().pop().Interface())
}
