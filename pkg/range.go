package pkg

import (
	"fmt"
	"go/token"
	"go/types"
	"reflect"
	"strconv"

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

func (r RangeStmt) eval(vm *VM) {} // noop

// Flow builds the control flow graph for the RangeStmt.
// Based on the Kind of X, it will branch into one of three flows:
//   - mapFlow for maps
//   - sliceOrArrayFlow for slices and arrays
//   - intFlow for integers
//   - funcFlow for functions of type iter.Seq[V any], iter.Seq2[K comparable, V any]
//   - chanFlow for channels
//
// All flows converge to a done step.
// Some subflows are transformed into a ForStmt that uses a hidden index variable.
func (r RangeStmt) flow(g *graphBuilder) (head Step) {
	head = r.x.flow(g)
	switcher := new(rangeIteratorSwitchStep)
	g.nextStep(switcher)
	// all flows converge to this done step
	rangeDone := g.newLabeledStep("~range-done", r.pos())

	// flow depends on the type of X
	switch r.xType.Underlying().(type) {
	case *types.Map:
		// start the map flow, detached from the current
		g.current = g.newLabeledStep("~range-map", r.pos())
		switcher.mapFlow = r.mapFlow(g)
		g.nextStep(rangeDone)

	case *types.Slice, *types.Array:
		// start the list flow, detached from the current
		g.current = g.newLabeledStep("~range-slice-or-array", r.pos())
		switcher.sliceOrArrayFlow = r.sliceOrArrayFlow(g)
		g.nextStep(rangeDone)

	case *types.Basic:
		basicKind := r.xType.Underlying().(*types.Basic).Kind()
		if basicKind == types.Int {
			// start the int flow, detached from the current
			g.current = g.newLabeledStep("~range-int", r.pos())
			switcher.intFlow = r.intFlow(g)
			g.nextStep(rangeDone)
			break
		}
		if basicKind == types.String || basicKind == types.UntypedString {
			// start the runes flow, detached from the current
			g.current = g.newLabeledStep("~range-runes", r.pos())
			switcher.sliceOrArrayFlow = r.sliceOrArrayFlow(g)
			g.nextStep(rangeDone)
			break
		}
		g.fatal(fmt.Sprintf("unhandled range over basic type %v", r.xType))
	case *types.Chan:
		// start the chan flow, detached from the current
		g.current = nil // g.newLabeledStep("~range-chan", r.pos())
		switcher.chanFlow = r.chanFlow(g)
		g.nextStep(rangeDone)
	default:
		g.fatal(fmt.Sprintf("unhandled range over type %v", r.xType))
	}
	return
}

type rangeMapIteratorInitStep struct {
	step
	localVarName string
	rangePos     token.Pos
}

func (r *rangeMapIteratorInitStep) take(vm *VM) Step {
	rangeable := vm.popOperand()
	iter := rangeable.MapRange()
	vm.currentEnv().set(r.localVarName, reflect.ValueOf(iter))
	return r.next
}

func (r *rangeMapIteratorInitStep) traverse(g *dot.Graph, fs *token.FileSet) dot.Node {
	return r.step.traverseWithLabel(g, r.step.StringWith("map-iterator-init"), sourceLocation(fs, r.pos()), fs)
}

func (r *rangeMapIteratorInitStep) pos() token.Pos {
	return r.rangePos
}

func (r *rangeMapIteratorInitStep) String() string {
	if r == nil {
		return "rangeMapIteratorInitStep(<nil>)"
	}
	return r.step.StringWith("~range-map-iterator-init:" + r.localVarName)
}

type rangeMapIteratorNextStep struct {
	step
	localVarName         string
	bodyFlow             Step
	yieldKey, yieldValue bool
	rangePos             token.Pos
}

func (r *rangeMapIteratorNextStep) take(vm *VM) Step {
	iterator := vm.currentEnv().valueLookUp(r.localVarName).Interface().(*reflect.MapIter)
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

func (r *rangeMapIteratorNextStep) traverse(g *dot.Graph, fs *token.FileSet) dot.Node {
	me := r.step.traverseWithLabel(g, r.step.StringWith("map-iterator-next"), sourceLocation(fs, r.pos()), fs)
	if r.bodyFlow != nil {
		// no edge if visited before
		sid := strconv.Itoa(r.bodyFlow.ID())
		if !g.HasNodeWithID(sid) {
			bodyNode := r.bodyFlow.traverse(g, fs)
			me.Edge(bodyNode, "body")
		}
	}
	return me
}

func (r *rangeMapIteratorNextStep) pos() token.Pos {
	return r.rangePos
}

func (r *rangeMapIteratorNextStep) String() string {
	if r == nil {
		return "rangeMapIteratorNextStep(<nil>)"
	}
	return r.step.StringWith("~range-map-iterator-next:" + r.localVarName)
}

func (r RangeStmt) chanFlow(g *graphBuilder) (head Step) {
	// <- chan
	recv := UnaryExpr{
		opPos: r.pos(),
		op:    token.ARROW,
		x:     r.x,
	}
	// var := <- chan
	ass := AssignStmt{
		tok:    token.DEFINE,
		tokPos: r.pos(),
		lhs:    []Expr{r.key},
		rhs:    []Expr{recv},
	}
	bodyList := append([]Stmt{ass}, r.body.list...)
	body := &BlockStmt{lbracePos: r.body.pos(), list: bodyList}
	return ForStmt{forPos: r.pos(), body: body}.flow(g)
}

func (r RangeStmt) mapFlow(g *graphBuilder) (head Step) {
	head = r.x.flow(g) // again on the stack so we can call MapRange to set the iterator

	// create the iterator
	localVarName := internalVarName("mapIter", g.idgen)
	init := new(rangeMapIteratorInitStep)
	init.rangePos = r.pos()
	init.localVarName = localVarName
	g.nextStep(init)

	// iterator next step
	iter := new(rangeMapIteratorNextStep)
	iter.rangePos = r.pos()
	iter.localVarName = localVarName
	iter.yieldKey = r.key != nil
	iter.yieldValue = r.value != nil
	g.nextStep(iter)

	// start the body flow, detached from the current
	g.current = g.newLabeledStep("~range-map-body", r.pos())
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
			tokPos: r.pos(),
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

func (r RangeStmt) intFlow(g *graphBuilder) (head Step) {

	// index := 0
	indexVar := Ident{name: internalVarName("index", g.idgen)}
	zeroInt := newBasicLit(r.pos(), reflect.ValueOf(0))
	initIndex := AssignStmt{
		tok:    token.DEFINE,
		tokPos: r.pos(),
		lhs:    []Expr{indexVar},
		rhs:    []Expr{zeroInt},
	}
	init := BlockStmt{list: []Stmt{initIndex}}

	// key := 0 // only one var permitted
	if r.key != nil {
		initKey := AssignStmt{
			tok:    token.DEFINE,
			tokPos: r.pos(),
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
		tokPos: r.pos(),
		x:      indexVar,
	}
	body := &BlockStmt{
		list: r.body.list,
	}
	// key = index
	if r.key != nil {
		updateKey := AssignStmt{
			tok:    token.ASSIGN,
			tokPos: r.pos(),
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

func (r RangeStmt) sliceOrArrayFlow(g *graphBuilder) (head Step) {

	// index := 0
	indexVar := Ident{name: internalVarName("index", g.idgen)}
	zeroInt := newBasicLit(r.pos(), reflect.ValueOf(0))
	initIndex := AssignStmt{
		tok:    token.DEFINE,
		tokPos: r.pos(),
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
			lbrackPos: r.pos(),
			x:         r.x,
			index:     indexVar,
		})
	}
	// key := x[0]
	// value := x[0]
	initKeyValue := AssignStmt{
		tok:    token.DEFINE,
		tokPos: r.pos(),
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
		tokPos: r.pos(),
		x:      indexVar,
	}
	// key = x[index]
	// value = x[index]
	updateKeyValue := AssignStmt{
		tok:    token.ASSIGN,
		tokPos: r.pos(),
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

func (r RangeStmt) pos() token.Pos { return r.forPos }

func (r RangeStmt) String() string {
	return fmt.Sprintf("RangeStmt(%v, %v, %v, %v)", r.key, r.value, r.x, r.body)
}

func (r RangeStmt) stmtStep() Evaluable { return r }

// rangeIteratorSwitchStep looks at the Kind of the value of X to determine which flow to use.
type rangeIteratorSwitchStep struct {
	step
	mapFlow          Step
	sliceOrArrayFlow Step
	intFlow          Step
	chanFlow         Step
}

func (i *rangeIteratorSwitchStep) take(vm *VM) Step {
	rangeable := vm.popOperand()
	if rangeable.Kind() == reflect.Pointer {
		rangeable = rangeable.Elem()
	}
	switch rangeable.Kind() {
	case reflect.Chan:
		return i.chanFlow.take(vm)
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
		vm.fatalf("cannot range over type %v", rangeable.Type())
	}
	return nil
}

func (i *rangeIteratorSwitchStep) traverse(g *dot.Graph, fs *token.FileSet) dot.Node {
	me := i.step.traverseWithLabel(g, i.step.StringWith("~range-iterator-switch"), sourceLocation(fs, i.pos()), fs)
	if i.mapFlow != nil {
		// no edge if visited before
		sid := strconv.Itoa(i.mapFlow.ID())
		if !g.HasNodeWithID(sid) {
			mapNode := i.mapFlow.traverse(g, fs)
			me.Edge(mapNode, "map")
		}
	}
	if i.sliceOrArrayFlow != nil {
		// no edge if visited before
		sid := strconv.Itoa(i.sliceOrArrayFlow.ID())
		if !g.HasNodeWithID(sid) {
			sliceOrArrayNode := i.sliceOrArrayFlow.traverse(g, fs)
			me.Edge(sliceOrArrayNode, "sliceOrArray")
		}
	}
	if i.intFlow != nil {
		// no edge if visited before
		sid := strconv.Itoa(i.intFlow.ID())
		if !g.HasNodeWithID(sid) {
			intNode := i.intFlow.traverse(g, fs)
			me.Edge(intNode, "int")
		}
	}
	if i.chanFlow != nil {
		// no edge if visited before
		sid := strconv.Itoa(i.chanFlow.ID())
		if !g.HasNodeWithID(sid) {
			chanNode := i.chanFlow.traverse(g, fs)
			me.Edge(chanNode, "chan")
		}
	}
	return me
}

type noExpr struct{}

func (noExpr) pos() token.Pos { return token.NoPos }
func (noExpr) eval(vm *VM)    {} // used?
func (n noExpr) flow(g *graphBuilder) (head Step) {
	return g.current
}
func (noExpr) String() string { return "NoExpr" }

type reflectLenExpr struct {
	X Expr
}

func (r reflectLenExpr) pos() token.Pos { return r.X.pos() }
func (r reflectLenExpr) eval(vm *VM) {
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
