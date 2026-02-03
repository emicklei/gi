package pkg

import (
	"fmt"
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
	forPos     token.Pos
	tok        token.Token // ILLEGAL if Key == nil, ASSIGN, DEFINE
	key, value Expr        // Key, Value may be nil
	x          Expr
	xType      types.Type // depending on type, different flows are created
	body       *BlockStmt
}

func (r RangeStmt) Eval(vm *VM) {} // noop

// Flow builds the control flow graph for the RangeStmt.
// Based on the Kind of X, it will branch into one of three flows:
//   - mapFlow for maps
//   - sliceOrArrayFlow for slices and arrays
//   - intFlow for integers
//   - funcFlow for functions of type iter.Seq[V any], iter.Seq2[K comparable, V any]
//
// All four flows converge to a done step.
// The last 3 subflows are transformed into a ForStmt that uses a hidden index variable.
func (r RangeStmt) flow(g *graphBuilder) (head Step) {
	head = r.x.flow(g)
	switcher := new(rangeIteratorSwitchStep)
	g.nextStep(switcher)
	// all flows converge to this done step
	rangeDone := g.newLabeledStep("range-done", r.Pos())

	// flow depends on the type of X
	switch r.xType.Underlying().(type) {
	case *types.Map:
		// start the map flow, detached from the current
		g.current = g.newLabeledStep("range-map", r.Pos())
		switcher.mapFlow = r.MapFlow(g)
		g.nextStep(rangeDone)

	case *types.Slice, *types.Array:
		// start the list flow, detached from the current
		g.current = g.newLabeledStep("range-slice-or-array", r.Pos())
		switcher.sliceOrArrayFlow = r.SliceOrArrayFlow(g)
		g.nextStep(rangeDone)

	case *types.Basic:
		basicKind := r.xType.Underlying().(*types.Basic).Kind()
		if basicKind == types.Int {
			// start the int flow, detached from the current
			g.current = g.newLabeledStep("range-int", r.Pos())
			switcher.intFlow = r.IntFlow(g)
			g.nextStep(rangeDone)
			break
		}
		if basicKind == types.String || basicKind == types.UntypedString {
			// start the runes flow, detached from the current
			g.current = g.newLabeledStep("range-runes", r.Pos())
			switcher.sliceOrArrayFlow = r.SliceOrArrayFlow(g)
			g.nextStep(rangeDone)
			break
		}
		g.fatal(fmt.Sprintf("unhandled range over basic type %v", r.xType))
	default:
		g.fatal(fmt.Sprintf("unhandled range over type %v", r.xType))
	}
	return
}

type rangeMapIteratorInitStep struct {
	step
	localVarName string
	pos          token.Pos
}

func (r *rangeMapIteratorInitStep) take(vm *VM) Step {
	rangeable := vm.popOperand()
	iter := rangeable.MapRange()
	vm.localEnv().set(r.localVarName, reflect.ValueOf(iter))
	return r.next
}

func (r *rangeMapIteratorInitStep) traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node {
	return r.step.traverseWithLabel(g, r.step.StringWith("map-iterator-init"), "next", visited)
}

func (r *rangeMapIteratorInitStep) Pos() token.Pos {
	return r.pos
}

func (r *rangeMapIteratorInitStep) String() string {
	if r == nil {
		return "rangeMapIteratorInitStep(<nil>)"
	}
	return r.step.StringWith("range-map-iterator-init:" + r.localVarName)
}

type rangeMapIteratorNextStep struct {
	step
	localVarName         string
	bodyFlow             Step
	yieldKey, yieldValue bool
	pos                  token.Pos
}

func (r *rangeMapIteratorNextStep) take(vm *VM) Step {
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

func (r *rangeMapIteratorNextStep) traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node {
	me := r.step.traverseWithLabel(g, r.step.StringWith("map-iterator-next"), "next", visited)
	if r.bodyFlow != nil {
		// no edge if visited before
		if _, ok := visited[r.bodyFlow.ID()]; !ok {
			bodyNode := r.bodyFlow.traverse(g, visited)
			me.Edge(bodyNode, "body")
		}
	}
	return me
}

func (r *rangeMapIteratorNextStep) Pos() token.Pos {
	return r.pos
}

func (r *rangeMapIteratorNextStep) String() string {
	if r == nil {
		return "rangeMapIteratorNextStep(<nil>)"
	}
	return r.step.StringWith("range-map-iterator-next:" + r.localVarName)
}

func (r RangeStmt) MapFlow(g *graphBuilder) (head Step) {
	head = r.x.flow(g) // again on the stack

	// create the iterator
	localVarName := internalVarName("mapIter", g.idgen)
	init := new(rangeMapIteratorInitStep)
	init.pos = r.Pos()
	init.localVarName = localVarName
	g.nextStep(init)

	// iterator next step
	iter := new(rangeMapIteratorNextStep)
	iter.pos = r.Pos()
	iter.localVarName = localVarName
	iter.yieldKey = r.key != nil
	iter.yieldValue = r.value != nil
	g.nextStep(iter)

	// start the body flow, detached from the current
	g.current = g.newLabeledStep("range-map-body", r.Pos())
	if iter.yieldKey || iter.yieldValue {
		// key = key
		// value = x[value]
		// value and key will be on the operand stack by the iterator step
		// so we use NoExpr as rhs placeholders
		lhs, rhs := []Expr{}, []Expr{}
		if iter.yieldKey {
			lhs = append(lhs, r.key)
			rhs = append(rhs, noExpr{}) // feels hacky
		}
		if iter.yieldValue {
			lhs = append(lhs, r.value)
			rhs = append(rhs, noExpr{}) // feels hacky
		}
		updateKeyValue := AssignStmt{
			tokPos: r.Pos(),
			tok:    token.DEFINE,
			lhs:    lhs,
			rhs:    rhs,
		}
		bodyFlow := updateKeyValue.flow(g)
		iter.bodyFlow = bodyFlow
		r.body.flow(g)
	} else {
		iter.bodyFlow = r.body.flow(g)
	}
	g.nextStep(iter) // back to iterator
	g.current = iter
	return
}

type noExpr struct{}

func (noExpr) Pos() token.Pos { return token.NoPos }
func (noExpr) Eval(vm *VM)    {} // used?
func (n noExpr) flow(g *graphBuilder) (head Step) {
	return g.current
}
func (noExpr) String() string { return "NoExpr" }

func (r RangeStmt) IntFlow(g *graphBuilder) (head Step) {

	// index := 0
	indexVar := Ident{name: internalVarName("index", g.idgen)}
	zeroInt := newBasicLit(r.Pos(), reflect.ValueOf(0))
	initIndex := AssignStmt{
		tok:    token.DEFINE,
		tokPos: r.Pos(),
		lhs:    []Expr{indexVar},
		rhs:    []Expr{zeroInt},
	}
	init := BlockStmt{list: []Stmt{initIndex}}

	// key := 0 // only one var permitted
	if r.key != nil {
		initKey := AssignStmt{
			tok:    token.DEFINE,
			tokPos: r.Pos(),
			lhs:    []Expr{r.key},
			rhs:    []Expr{indexVar},
		}
		init.list = append(init.list, initKey)
	}
	// index < x
	cond := BinaryExpr{
		op:    token.LSS,
		opPos: r.forPos,
		x:     indexVar,
		y:     r.x,
	}
	// index++
	post := IncDecStmt{
		tok:    token.INC,
		tokPos: r.Pos(),
		x:      indexVar,
	}
	body := &BlockStmt{
		list: r.body.list,
	}
	// key = index
	if r.key != nil {
		updateKey := AssignStmt{
			tok:    token.ASSIGN,
			tokPos: r.Pos(),
			lhs:    []Expr{r.key},
			rhs:    []Expr{indexVar},
		}
		// body with updated key assignment at the top
		body.list = append([]Stmt{updateKey}, body.list...)
	}
	// now build it
	forstmt := ForStmt{
		init: init,
		cond: cond,
		post: post,
		body: body,
	}
	return forstmt.flow(g)
}

func (r RangeStmt) SliceOrArrayFlow(g *graphBuilder) (head Step) {

	// index := 0
	indexVar := Ident{name: internalVarName("index", g.idgen)}
	zeroInt := newBasicLit(r.Pos(), reflect.ValueOf(0))
	initIndex := AssignStmt{
		tok:    token.DEFINE,
		tokPos: r.Pos(),
		lhs:    []Expr{indexVar},
		rhs:    []Expr{zeroInt},
	}
	// key and value are optionally defined
	lhs, rhs := []Expr{}, []Expr{}
	if r.key != nil {
		lhs = append(lhs, r.key)
		rhs = append(rhs, indexVar)
	}
	if r.value != nil {
		lhs = append(lhs, r.value)
		rhs = append(rhs, IndexExpr{
			lbrackPos: r.Pos(),
			x:         r.x,
			index:     indexVar,
		})
	}
	// key := x[0]
	// value := x[0]
	initKeyValue := AssignStmt{
		tok:    token.DEFINE,
		tokPos: r.Pos(),
		lhs:    lhs,
		rhs:    rhs,
	}
	init := BlockStmt{
		list: []Stmt{
			initIndex,
			initKeyValue,
		},
	}
	// index < len(x)
	cond := BinaryExpr{
		op:    token.LSS,
		opPos: r.forPos,
		x:     indexVar,
		y:     reflectLenExpr{X: r.x},
	}
	// index++
	post := IncDecStmt{
		tok:    token.INC,
		tokPos: r.Pos(),
		x:      indexVar,
	}
	// key = x[index]
	// value = x[index]
	updateKeyValue := AssignStmt{
		tok:    token.ASSIGN,
		tokPos: r.Pos(),
		lhs:    lhs,
		rhs:    rhs,
	}
	// body with updated key/value assignment at the top
	body := &BlockStmt{
		list: append([]Stmt{updateKeyValue}, r.body.list...),
	}
	// now build it
	forstmt := ForStmt{
		init: init,
		cond: cond,
		post: post,
		body: body,
	}
	return forstmt.flow(g)
}

func (r RangeStmt) Pos() token.Pos { return r.forPos }

func (r RangeStmt) String() string {
	return fmt.Sprintf("RangeStmt(%v, %v, %v, %v)", r.key, r.value, r.x, r.body)
}

func (r RangeStmt) stmtStep() Evaluable { return r }

type reflectLenExpr struct {
	X Expr
}

func (r reflectLenExpr) Pos() token.Pos { return r.X.Pos() }
func (r reflectLenExpr) Eval(vm *VM) {
	val := vm.popOperand()
	vm.pushOperand(reflect.ValueOf(val.Len()))
}
func (r reflectLenExpr) flow(g *graphBuilder) (head Step) {
	head = r.X.flow(g)
	g.next(r)
	return
}
func (r reflectLenExpr) String() string {
	return fmt.Sprintf("reflectLenExpr(%v)", r.X)
}

// rangeIteratorSwitchStep looks at the Kind of the value of X to determine which flow to use.
type rangeIteratorSwitchStep struct {
	step
	mapFlow          Step
	sliceOrArrayFlow Step
	intFlow          Step
}

func (i *rangeIteratorSwitchStep) take(vm *VM) Step {
	rangeable := vm.popOperand()
	if rangeable.Kind() == reflect.Ptr {
		rangeable = rangeable.Elem()
	}
	switch rangeable.Kind() {
	case reflect.Map:
		return i.mapFlow.take(vm)
	case reflect.Slice, reflect.Array:
		return i.sliceOrArrayFlow.take(vm)
	case reflect.Int:
		return i.intFlow.take(vm)
	case reflect.String:
		// iterate over runes
		str := rangeable.String()
		runeSlice := make([]rune, 0, len(str))
		for _, r := range str {
			runeSlice = append(runeSlice, r)
		}
		vm.pushOperand(reflect.ValueOf(runeSlice))
		return i.sliceOrArrayFlow.take(vm)
	case reflect.Func:
		// TODO
		// func(yield func(V) bool): //iter.Seq[V any]:
		//return i.funcFlow.take(vm)
		fallthrough
	default:
		panic(fmt.Sprintf("cannot range over type %v", rangeable.Type()))
	}
}
func (i *rangeIteratorSwitchStep) traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node {
	me := i.step.traverseWithLabel(g, i.step.StringWith("switch-iterator"), "next", visited)
	if i.mapFlow != nil {
		// no edge if visited before
		if _, ok := visited[i.mapFlow.ID()]; !ok {
			mapNode := i.mapFlow.traverse(g, visited)
			me.Edge(mapNode, "map")
		}
	}
	if i.sliceOrArrayFlow != nil {
		// no edge if visited before
		if _, ok := visited[i.sliceOrArrayFlow.ID()]; !ok {
			sliceOrArrayNode := i.sliceOrArrayFlow.traverse(g, visited)
			me.Edge(sliceOrArrayNode, "sliceOrArray")
		}
	}
	if i.intFlow != nil {
		// no edge if visited before
		if _, ok := visited[i.intFlow.ID()]; !ok {
			intNode := i.intFlow.traverse(g, visited)
			me.Edge(intNode, "int")
		}
	}
	return me
}
