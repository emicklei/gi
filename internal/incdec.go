package internal

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
)

var _ Flowable = IncDecStmt{}
var _ Stmt = IncDecStmt{}

type IncDecStmt struct {
	*ast.IncDecStmt
	X Expr
}

func (i IncDecStmt) Flow(g *graphBuilder) (head Step) {
	head = i.X.Flow(g)
	g.next(i)
	return head
}

func (i IncDecStmt) Eval(vm *VM) {
	val := vm.frameStack.top().pop()
	if !val.IsValid() {
		// TODO
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

func (i IncDecStmt) stmtStep() Evaluable { return i }

func (i IncDecStmt) String() string {
	return fmt.Sprintf("IncDecStmt(%v)", i.X)
}
