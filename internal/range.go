package internal

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"

	"github.com/emicklei/dot"
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
	switch rangeable.Kind() {
	case reflect.Map:
		iter := rangeable.MapRange()
		for iter.Next() {
			vm.pushNewFrame(r)
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
			if trace {
				vm.eval(r.Body)
			} else {
				r.Body.Eval(vm)
			}
			vm.popFrame()
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < rangeable.Len(); i++ {
			vm.pushNewFrame(r)
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
			if trace {
				vm.eval(r.Body)
			} else {
				r.Body.Eval(vm)
			}
			vm.popFrame()
		}
	case reflect.Int:
		for i := 0; i < int(rangeable.Int()); i++ {
			vm.pushNewFrame(r)
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
			if trace {
				vm.eval(r.Body)
			} else {
				r.Body.Eval(vm)
			}
			vm.popFrame()
		}
	}
}

// Flow builds the control flow graph for the RangeStmt.
// Based on the Kind of X, it will branch into one of three flows:
//   - mapFlow for maps
//   - sliceOrArrayFlow for slices and arrays
//   - intFlow for integers
//
// All three flows converge to a done step.
// Each subflow is transformed into a ForStmt that uses an hidden index variable.
// TODO fix position info
func (r RangeStmt) Flow(g *graphBuilder) (head Step) {
	head = r.X.Flow(g)
	switcher := &rangeIteratorSwitchStep{
		step: newStep(nil),
	}
	g.nextStep(switcher)
	// all flows converge to this done step
	rangeDone := newStep(nil)

	// start the map flow, detached from the current
	g.current = newStep(nil)
	switcher.mapFlow = r.MapFlow(g)
	g.nextStep(rangeDone)

	// start the list flow, detached from the current
	g.current = newStep(nil)
	switcher.sliceOrArrayFlow = r.SliceOrArrayFlow(g)
	g.nextStep(rangeDone)

	// start the int flow, detached from the current
	g.current = newStep(nil)
	switcher.intFlow = r.IntFlow(g)
	g.nextStep(rangeDone)
	return
}

type rangeMapIteratorStep struct {
	*step
	iterator *reflect.MapIter
	bodyFlow Step
}

func (r *rangeMapIteratorStep) Take(vm *VM) Step {
	if r.iterator == nil {
		rangeable := vm.callStack.top().pop()
		r.iterator = rangeable.MapRange()
	}
	if r.iterator.Next() {
		// first value then key to match assignment order
		vm.callStack.top().push(reflect.ValueOf(r.iterator.Value()))
		vm.callStack.top().push(reflect.ValueOf(r.iterator.Key()))
		return r.bodyFlow
	}
	return r.next
}

func (r *rangeMapIteratorStep) Traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node {
	me := r.step.traverse(g, r.step.StringWith("map-iterator"), "next", visited)
	if r.bodyFlow != nil {
		// no edge if visited before
		if _, ok := visited[r.bodyFlow.ID()]; !ok {
			bodyNode := r.bodyFlow.Traverse(g, visited)
			me.Edge(bodyNode, "body")
		}
	}
	return me
}

func (r RangeStmt) MapFlow(g *graphBuilder) (head Step) {
	head = r.X.Flow(g) // again on the stack
	iter := &rangeMapIteratorStep{
		step: newStep(nil),
	}
	g.nextStep(iter)

	// start the body flow, detached from the current
	g.current = newStep(nil)
	bodyFlow := g.newPushStackFrame()
	g.nextStep(bodyFlow)
	// key = key
	// value = x[value]
	// value and key are on the operand stack
	updateKeyValue := AssignStmt{
		AssignStmt: &ast.AssignStmt{
			Tok: token.ASSIGN,
		},
		Lhs: []Expr{r.Key, r.Value},
		Rhs: []Expr{NoExpr{}, NoExpr{}},
	}
	updateKeyValue.Flow(g)
	r.Body.Flow(g)
	g.nextStep(iter) // back to iterator
	iter.bodyFlow = bodyFlow

	g.current = iter
	return
}

type NoExpr struct{}

func (NoExpr) Eval(vm *VM) {}
func (n NoExpr) Flow(g *graphBuilder) (head Step) {
	g.next(n)
	return g.current
}

func (r RangeStmt) IntFlow(g *graphBuilder) (head Step) {

	// index := 0
	indexVar := Ident{Ident: &ast.Ident{Name: fmt.Sprintf("_index_%d", idgen)}} // must be unique in env
	zeroInt := BasicLit{BasicLit: &ast.BasicLit{Kind: token.INT, Value: "0"}}
	initIndex := AssignStmt{
		AssignStmt: &ast.AssignStmt{
			Tok: token.DEFINE,
		},
		Lhs: []Expr{indexVar},
		Rhs: []Expr{zeroInt},
	}
	init := BlockStmt{
		List: []Stmt{
			initIndex,
		},
	}
	// key := 0 // only one var permitted
	if r.Key != nil {
		initKey := AssignStmt{
			AssignStmt: &ast.AssignStmt{
				Tok: token.DEFINE,
			},
			Lhs: []Expr{r.Key},
			Rhs: []Expr{indexVar},
		}
		init.List = append(init.List, initKey)
	}
	// index < x
	cond := BinaryExpr{
		BinaryExpr: &ast.BinaryExpr{
			Op: token.LSS,
		},
		X: indexVar,
		Y: r.X,
	}
	// index++
	post := IncDecStmt{
		IncDecStmt: &ast.IncDecStmt{
			Tok: token.INC,
		},
		X: indexVar,
	}
	body := &BlockStmt{
		List: r.Body.List,
	}
	// key = index
	if r.Key != nil {
		updateKey := AssignStmt{
			AssignStmt: &ast.AssignStmt{
				Tok: token.ASSIGN,
			},
			Lhs: []Expr{r.Key},
			Rhs: []Expr{indexVar},
		}
		// body with updated key assignment at the top
		body.List = append([]Stmt{updateKey}, body.List...)
	}
	// now build it
	forstmt := ForStmt{
		Init: init,
		Cond: cond,
		Post: post,
		Body: body,
	}
	return forstmt.Flow(g)
}

func (r RangeStmt) SliceOrArrayFlow(g *graphBuilder) (head Step) {

	// index := 0
	indexVar := Ident{Ident: &ast.Ident{Name: fmt.Sprintf("_index_%d", idgen)}} // must be unique in env
	zeroInt := BasicLit{BasicLit: &ast.BasicLit{Kind: token.INT, Value: "0"}}
	initIndex := AssignStmt{
		AssignStmt: &ast.AssignStmt{
			Tok: token.DEFINE,
		},
		Lhs: []Expr{indexVar},
		Rhs: []Expr{zeroInt},
	}
	// key := x[0]
	// value := x[0]
	initKeyValue := AssignStmt{
		AssignStmt: &ast.AssignStmt{
			Tok: token.DEFINE,
		},
		Lhs: []Expr{r.Key, r.Value},
		Rhs: []Expr{indexVar, IndexExpr{
			X:     r.X,
			Index: indexVar,
		}},
	}
	init := BlockStmt{
		List: []Stmt{
			initIndex,
			initKeyValue,
		},
	}
	// index < len(x)
	cond := BinaryExpr{
		BinaryExpr: &ast.BinaryExpr{
			Op: token.LSS,
		},
		X: indexVar,
		Y: ReflectLenExpr{X: r.X},
	}
	// index++
	post := IncDecStmt{
		IncDecStmt: &ast.IncDecStmt{
			Tok: token.INC,
		},
		X: indexVar,
	}
	// key = x[index]
	// value = x[index]
	updateKeyValue := AssignStmt{
		AssignStmt: &ast.AssignStmt{
			Tok: token.ASSIGN,
		},
		Lhs: []Expr{r.Key, r.Value},
		Rhs: []Expr{indexVar, IndexExpr{
			X:     r.X,
			Index: indexVar,
		}},
	}
	// body with updated key/value assignment at the top
	body := &BlockStmt{
		List: append([]Stmt{updateKeyValue}, r.Body.List...),
	}
	// now build it
	forstmt := ForStmt{
		Init: init,
		Cond: cond,
		Post: post,
		Body: body,
	}
	return forstmt.Flow(g)
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

// rangeIteratorSwitchStep looks at the Kind of the value of X to determine which flow to use.
type rangeIteratorSwitchStep struct {
	*step
	mapFlow          Step
	sliceOrArrayFlow Step
	intFlow          Step
}

func (i *rangeIteratorSwitchStep) Take(vm *VM) Step {
	rangeable := vm.callStack.top().pop()
	switch rangeable.Kind() {
	case reflect.Map:
		return i.mapFlow.Take(vm)
	case reflect.Slice, reflect.Array:
		return i.sliceOrArrayFlow.Take(vm)
	case reflect.Int:
		return i.intFlow.Take(vm)
	default:
		panic(fmt.Sprintf("cannot range over type %v", rangeable.Type()))
	}
}
func (i *rangeIteratorSwitchStep) Traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node {
	me := i.step.traverse(g, i.step.StringWith("switch-iterator"), "next", visited)
	if i.mapFlow != nil {
		// no edge if visited before
		if _, ok := visited[i.mapFlow.ID()]; !ok {
			mapNode := i.mapFlow.Traverse(g, visited)
			me.Edge(mapNode, "map")
		}
	}
	if i.sliceOrArrayFlow != nil {
		// no edge if visited before
		if _, ok := visited[i.sliceOrArrayFlow.ID()]; !ok {
			sliceOrArrayNode := i.sliceOrArrayFlow.Traverse(g, visited)
			me.Edge(sliceOrArrayNode, "sliceOrArray")
		}
	}
	if i.intFlow != nil {
		// no edge if visited before
		if _, ok := visited[i.intFlow.ID()]; !ok {
			intNode := i.intFlow.Traverse(g, visited)
			me.Edge(intNode, "int")
		}
	}
	return me
}
