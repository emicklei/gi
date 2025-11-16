package internal

import (
	"fmt"
	"go/token"
)

var _ Stmt = BlockStmt{}

type BlockStmt struct {
	LbracePos token.Pos // position of "{"
	List      []Stmt
}

func (b BlockStmt) Eval(vm *VM) {
	for _, stmt := range b.List {
		vm.eval(stmt.stmtStep())
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

func (b BlockStmt) stmtStep() Evaluable { return b }

func (b BlockStmt) Pos() token.Pos { return b.LbracePos }

func (b BlockStmt) String() string {
	return fmt.Sprintf("BlockStmt(len=%d)", len(b.List))
}
