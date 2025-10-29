package internal

import (
	"fmt"
	"go/ast"
)

var _ Stmt = BlockStmt{}

type BlockStmt struct {
	*ast.BlockStmt
	List []Stmt
}

func (b BlockStmt) stmtStep() Evaluable { return b }

func (b BlockStmt) String() string {
	return fmt.Sprintf("BlockStmt(len=%d)", len(b.List))
}

func (b BlockStmt) Eval(vm *VM) {
	for _, stmt := range b.List {
		if trace {
			vm.traceEval(stmt.stmtStep())
		} else {
			stmt.stmtStep().Eval(vm)
		}
	}
}

func (b BlockStmt) Flow(g *graphBuilder) (head Step) {
	head = g.current
	for i, stmt := range b.List {
		if i == 0 {
			head = stmt.Flow(g)
			continue
		}
		_ = stmt.Flow(g)
	}
	return head
}
