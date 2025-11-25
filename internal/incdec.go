package internal

import (
	"fmt"
	"go/token"
	"reflect"
)

type IncDecFunc func(v reflect.Value) reflect.Value

var _ Flowable = IncDecStmt{}
var _ Stmt = IncDecStmt{}

type IncDecStmt struct {
	TokPos token.Pos   // position of Tok
	Tok    token.Token // INC or DEC
	X      Expr
}

func (i IncDecStmt) Flow(g *graphBuilder) (head Step) {
	head = i.X.Flow(g)
	g.next(i)
	return head
}

func (i IncDecStmt) Eval(vm *VM) {
	val := vm.callStack.top().pop()
	// propagate undeclared value. this happens when the expression is
	// used in a package variable or constant declaration
	if val == reflectUndeclared {
		return
	}
	if i.Tok == token.INC {
		switch val.Kind() {
		case reflect.Int:
			if a, ok := i.X.(CanAssign); ok {
				a.Assign(vm, reflect.ValueOf(int(val.Int()+1)))
			}
		case reflect.Int32:
			if a, ok := i.X.(CanAssign); ok {
				a.Assign(vm, reflect.ValueOf(int32(val.Int()+1)))
			}
		case reflect.Int64:
			if a, ok := i.X.(CanAssign); ok {
				a.Assign(vm, reflect.ValueOf(int64(val.Int()+1)))
			}
		case reflect.Float32:
			if a, ok := i.X.(CanAssign); ok {
				a.Assign(vm, reflect.ValueOf(float32(val.Float()+1)))
			}
		case reflect.Float64:
			if a, ok := i.X.(CanAssign); ok {
				a.Assign(vm, reflect.ValueOf(val.Float()+1))
			}
		default:
			panic("unsupported type for ++: " + val.Kind().String())
		}
	} else { // DEC
		switch val.Kind() {
		case reflect.Int:
			if a, ok := i.X.(CanAssign); ok {
				a.Assign(vm, reflect.ValueOf(int(val.Int()-1)))
			}
		case reflect.Int32:
			if a, ok := i.X.(CanAssign); ok {
				a.Assign(vm, reflect.ValueOf(int32(val.Int()-1)))
			}
		case reflect.Int64:
			if a, ok := i.X.(CanAssign); ok {
				a.Assign(vm, reflect.ValueOf(val.Int()-1))
			}
		case reflect.Float64:
			if a, ok := i.X.(CanAssign); ok {
				a.Assign(vm, reflect.ValueOf(val.Float()-1))
			}
		case reflect.Float32:
			if a, ok := i.X.(CanAssign); ok {
				a.Assign(vm, reflect.ValueOf(float32(val.Float()-1)))
			}
		default:
			panic("unsupported type for -- :" + val.Kind().String())
		}
	}
}
func (i IncDecStmt) Pos() token.Pos { return i.TokPos }

func (i IncDecStmt) stmtStep() Evaluable { return i }

func (i IncDecStmt) String() string {
	return fmt.Sprintf("IncDecStmt(%v)", i.X)
}
