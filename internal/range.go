package internal

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
)

var (
	_ Flowable = RangeStmt{}
	_ Stmt     = RangeStmt{}
)

type RangeStmt struct {
	*ast.RangeStmt
	Key, Value Expr // Key, Value may be nil
	X          Expr
	Body       *BlockStmt
}

func (r RangeStmt) Eval(vm *VM) {
	rangeable := vm.returnsEval(r.X)
	vm.pushNewFrame(r)

	// special case for Map
	if rangeable.Kind() == reflect.Map {
		iter := rangeable.MapRange()
		for iter.Next() {
			if r.Key != nil {
				if ca, ok := r.Key.(CanAssign); ok {
					ca.Define(vm, iter.Key())
				}
			}
			if r.Value != nil {
				if ca, ok := r.Value.(CanAssign); ok {
					ca.Define(vm, iter.Value())
				}
			}
			vm.eval(r.Body)
		}
	}
	if rangeable.Kind() == reflect.Slice || rangeable.Kind() == reflect.Array {
		for i := 0; i < rangeable.Len(); i++ {
			if r.Key != nil {
				if ca, ok := r.Key.(CanAssign); ok {
					ca.Define(vm, reflect.ValueOf(i))
				}
			}
			if r.Value != nil {
				if ca, ok := r.Value.(CanAssign); ok {
					ca.Define(vm, rangeable.Index(i))
				}
			}
			vm.eval(r.Body)
		}
	}
	vm.popFrame()
}

func (r RangeStmt) Flow(g *graphBuilder) (head Step) {
	head = r.X.Flow(g)
	push := g.newPushStackFrame()
	g.nextStep(push)

	// index := 0
	indexVar := Ident{Ident: &ast.Ident{Name: fmt.Sprintf("_index_%d", idgen)}} // must be unique in env
	zeroInt := BasicLit{BasicLit: &ast.BasicLit{Kind: token.INT, Value: "0"}}
	assign := AssignStmt{
		AssignStmt: &ast.AssignStmt{
			Tok: token.DEFINE,
		},
		Lhs: []Expr{indexVar},
		Rhs: []Expr{zeroInt},
	}
	assign.Flow(g)

	// index < len(x)
	condition := BinaryExpr{
		BinaryExpr: &ast.BinaryExpr{
			Op: token.LSS,
		},
		X: indexVar,
		Y: ReflectLenExpr{X: r.X},
	}
	condition.Flow(g)

	// index++
	indexInc := IncDecStmt{
		IncDecStmt: &ast.IncDecStmt{
			Tok: token.INC,
		},
		X: indexVar,
	}
	indexInc.Flow(g)

	pop := g.newPopStackFrame()
	g.nextStep(pop)
	return
}

func (r RangeStmt) String() string {
	return fmt.Sprintf("RangeStmt(%v, %v, %v, %v)", r.Key, r.Value, r.X, r.Body)
}

func (r RangeStmt) stmtStep() Evaluable { return r }

type ReflectLenExpr struct {
	// TODO position info
	X Expr
}

func (r ReflectLenExpr) Eval(vm *VM) {
	val := vm.callStack.top().pop()
	vm.callStack.top().push(reflect.ValueOf(val.Len()))
}
func (r ReflectLenExpr) Flow(g *graphBuilder) (head Step) {
	head = r.X.Flow(g)
	g.next(r)
	return
}
func (r ReflectLenExpr) String() string {
	return fmt.Sprintf("ReflectLenExpr(%v)", r.X)
}
