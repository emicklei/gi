package internal

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
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
				vm.traceEval(r.Body)
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
				vm.traceEval(r.Body)
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
				vm.traceEval(r.Body)
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
// The last 2 subflows are transformed into a ForStmt that uses a hidden index variable.
// TODO fix position info
func (r RangeStmt) Flow(g *graphBuilder) (head Step) {
	head = r.X.Flow(g)
	switcher := &rangeIteratorSwitchStep{
		step: newStep(nil),
	}
	g.nextStep(switcher)
	// all flows converge to this done step
	rangeDone := newStep(nil)

	// determine the type of X
	goType := g.goPkg.TypesInfo.TypeOf(r.RangeStmt.X)

	switch goType.Underlying().(type) {
	case *types.Map:
		// start the map flow, detached from the current
		g.current = newStep(nil)
		switcher.mapFlow = r.MapFlow(g)
		g.nextStep(rangeDone)

	case *types.Slice, *types.Array:
		// start the list flow, detached from the current
		g.current = newStep(nil)
		switcher.sliceOrArrayFlow = r.SliceOrArrayFlow(g)
		g.nextStep(rangeDone)

	case *types.Basic:
		// start the int flow, detached from the current
		g.current = newStep(nil)
		switcher.intFlow = r.IntFlow(g)
		g.nextStep(rangeDone)
	}
	return
}

type rangeMapIteratorInitStep struct {
	*step
	localVarName string
}

func (r *rangeMapIteratorInitStep) Take(vm *VM) Step {
	rangeable := vm.callStack.top().pop()
	iter := rangeable.MapRange()
	vm.localEnv().set(r.localVarName, reflect.ValueOf(iter))
	return r.next
}

func (r *rangeMapIteratorInitStep) Traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node {
	return r.step.traverse(g, r.step.StringWith("map-iterator-init"), "next", visited)
}

func (r *rangeMapIteratorInitStep) String() string {
	return r.step.StringWith("range-map-iterator-init:" + r.localVarName)
}

type rangeMapIteratorNextStep struct {
	*step
	localVarName         string
	bodyFlow             Step
	yieldKey, yieldValue bool
}

func (r *rangeMapIteratorNextStep) Take(vm *VM) Step {
	iterator := vm.localEnv().valueLookUp(r.localVarName).Interface().(*reflect.MapIter)
	if iterator.Next() {
		// first value then key to match assignment order
		if r.yieldValue {
			vm.pushOperand(iterator.Value())
		}
		if r.yieldKey {
			vm.pushOperand(iterator.Key())
		}
		return r.bodyFlow
	}
	return r.next
}

func (r *rangeMapIteratorNextStep) Traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node {
	me := r.step.traverse(g, r.step.StringWith("map-iterator-next"), "next", visited)
	if r.bodyFlow != nil {
		// no edge if visited before
		if _, ok := visited[r.bodyFlow.ID()]; !ok {
			bodyNode := r.bodyFlow.Traverse(g, visited)
			me.Edge(bodyNode, "body")
		}
	}
	return me
}

func (r *rangeMapIteratorNextStep) String() string {
	return r.step.StringWith("range-map-iterator-next:" + r.localVarName)
}

func (r RangeStmt) MapFlow(g *graphBuilder) (head Step) {
	head = r.X.Flow(g) // again on the stack

	// create the iterator
	localVarName := fmt.Sprintf("_mapIter_%d", idgen)
	init := &rangeMapIteratorInitStep{
		step:         newStep(nil),
		localVarName: localVarName,
	}
	g.nextStep(init)

	// iterator next step
	iter := &rangeMapIteratorNextStep{
		step:         newStep(nil),
		localVarName: localVarName,
		yieldKey:     r.Key != nil,
		yieldValue:   r.Value != nil,
	}
	g.nextStep(iter)

	// start the body flow, detached from the current
	g.current = newStep(nil)
	if iter.yieldKey || iter.yieldValue {
		// key = key
		// value = x[value]
		// value and key are on the operand stack by the iterator step
		lhs, rhs := []Expr{}, []Expr{}
		if iter.yieldKey {
			lhs = append(lhs, r.Key)
			rhs = append(rhs, NoExpr{}) // feels hacky
		}
		if iter.yieldValue {
			lhs = append(lhs, r.Value)
			rhs = append(rhs, NoExpr{}) // feels hacky
		}
		updateKeyValue := AssignStmt{
			AssignStmt: &ast.AssignStmt{
				Tok: token.DEFINE,
			},
			Lhs: lhs,
			Rhs: rhs,
		}
		bodyFlow := updateKeyValue.Flow(g)
		iter.bodyFlow = bodyFlow
		r.Body.Flow(g)
	} else {
		iter.bodyFlow = r.Body.Flow(g)
	}
	g.nextStep(iter) // back to iterator
	g.current = iter
	return
}

type NoExpr struct{}

func (NoExpr) Eval(vm *VM) {} // used?
func (n NoExpr) Flow(g *graphBuilder) (head Step) {
	return g.current
}
func (NoExpr) String() string { return "NoExpr" }

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
	init := BlockStmt{List: []Stmt{initIndex}}

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
	// key and value are optionally defined
	lhs, rhs := []Expr{}, []Expr{}
	if r.Key != nil {
		lhs = append(lhs, r.Key)
		rhs = append(rhs, indexVar)
	}
	if r.Value != nil {
		lhs = append(lhs, r.Value)
		rhs = append(rhs, IndexExpr{
			X:     r.X,
			Index: indexVar,
		})
	}
	// key := x[0]
	// value := x[0]
	initKeyValue := AssignStmt{
		AssignStmt: &ast.AssignStmt{
			Tok: token.DEFINE,
		},
		Lhs: lhs,
		Rhs: rhs,
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
		Lhs: lhs,
		Rhs: rhs,
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
