package internal

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
	ForPos     token.Pos
	Tok        token.Token // ILLEGAL if Key == nil, ASSIGN, DEFINE
	Key, Value Expr        // Key, Value may be nil
	X          Expr
	XType      types.Type // depending on type, different flows are created
	Body       *BlockStmt
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
func (r RangeStmt) Flow(g *graphBuilder) (head Step) {
	head = r.X.Flow(g)
	switcher := new(rangeIteratorSwitchStep)
	g.nextStep(switcher)
	// all flows converge to this done step
	rangeDone := g.newLabeledStep("range-done", r.Pos())

	// flow depends on the type of X
	switch r.XType.Underlying().(type) {
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
		basicKind := r.XType.Underlying().(*types.Basic).Kind()
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
		g.fatal(fmt.Sprintf("unhandled range over basic type %v", r.XType))
	default:
		g.fatal(fmt.Sprintf("unhandled range over type %v", r.XType))
	}
	return
}

type rangeMapIteratorInitStep struct {
	step
	localVarName string
	pos          token.Pos
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
	head = r.X.Flow(g) // again on the stack

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
	iter.yieldKey = r.Key != nil
	iter.yieldValue = r.Value != nil
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
			lhs = append(lhs, r.Key)
			rhs = append(rhs, noExpr{}) // feels hacky
		}
		if iter.yieldValue {
			lhs = append(lhs, r.Value)
			rhs = append(rhs, noExpr{}) // feels hacky
		}
		updateKeyValue := AssignStmt{
			TokPos: r.Pos(),
			Tok:    token.DEFINE,
			Lhs:    lhs,
			Rhs:    rhs,
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

type noExpr struct{}

func (noExpr) Pos() token.Pos { return token.NoPos }
func (noExpr) Eval(vm *VM)    {} // used?
func (n noExpr) Flow(g *graphBuilder) (head Step) {
	return g.current
}
func (noExpr) String() string { return "NoExpr" }

func (r RangeStmt) IntFlow(g *graphBuilder) (head Step) {

	// index := 0
	indexVar := Ident{Name: internalVarName("index", g.idgen)}
	zeroInt := newBasicLit(r.Pos(), reflect.ValueOf(0))
	initIndex := AssignStmt{
		Tok:    token.DEFINE,
		TokPos: r.Pos(),
		Lhs:    []Expr{indexVar},
		Rhs:    []Expr{zeroInt},
	}
	init := BlockStmt{List: []Stmt{initIndex}}

	// key := 0 // only one var permitted
	if r.Key != nil {
		initKey := AssignStmt{
			Tok:    token.DEFINE,
			TokPos: r.Pos(),
			Lhs:    []Expr{r.Key},
			Rhs:    []Expr{indexVar},
		}
		init.List = append(init.List, initKey)
	}
	// index < x
	cond := BinaryExpr{
		Op:    token.LSS,
		OpPos: r.ForPos,
		X:     indexVar,
		Y:     r.X,
	}
	// index++
	post := IncDecStmt{
		Tok:    token.INC,
		TokPos: r.Pos(),
		X:      indexVar,
	}
	body := &BlockStmt{
		List: r.Body.List,
	}
	// key = index
	if r.Key != nil {
		updateKey := AssignStmt{
			Tok:    token.ASSIGN,
			TokPos: r.Pos(),
			Lhs:    []Expr{r.Key},
			Rhs:    []Expr{indexVar},
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
	indexVar := Ident{Name: internalVarName("index", g.idgen)}
	zeroInt := newBasicLit(r.Pos(), reflect.ValueOf(0))
	initIndex := AssignStmt{
		Tok:    token.DEFINE,
		TokPos: r.Pos(),
		Lhs:    []Expr{indexVar},
		Rhs:    []Expr{zeroInt},
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
			Lbrack: r.Pos(),
			X:      r.X,
			Index:  indexVar,
		})
	}
	// key := x[0]
	// value := x[0]
	initKeyValue := AssignStmt{
		Tok:    token.DEFINE,
		TokPos: r.Pos(),
		Lhs:    lhs,
		Rhs:    rhs,
	}
	init := BlockStmt{
		List: []Stmt{
			initIndex,
			initKeyValue,
		},
	}
	// index < len(x)
	cond := BinaryExpr{
		Op:    token.LSS,
		OpPos: r.ForPos,
		X:     indexVar,
		Y:     reflectLenExpr{X: r.X},
	}
	// index++
	post := IncDecStmt{
		Tok:    token.INC,
		TokPos: r.Pos(),
		X:      indexVar,
	}
	// key = x[index]
	// value = x[index]
	updateKeyValue := AssignStmt{
		Tok:    token.ASSIGN,
		TokPos: r.Pos(),
		Lhs:    lhs,
		Rhs:    rhs,
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

func (r RangeStmt) Pos() token.Pos { return r.ForPos }

func (r RangeStmt) String() string {
	return fmt.Sprintf("RangeStmt(%v, %v, %v, %v)", r.Key, r.Value, r.X, r.Body)
}

func (r RangeStmt) stmtStep() Evaluable { return r }

type reflectLenExpr struct {
	X Expr
}

func (r reflectLenExpr) Pos() token.Pos { return r.X.Pos() }
func (r reflectLenExpr) Eval(vm *VM) {
	val := vm.callStack.top().pop()
	vm.callStack.top().push(reflect.ValueOf(val.Len()))
}
func (r reflectLenExpr) Flow(g *graphBuilder) (head Step) {
	head = r.X.Flow(g)
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

func (i *rangeIteratorSwitchStep) Take(vm *VM) Step {
	rangeable := vm.callStack.top().pop()
	if rangeable.Kind() == reflect.Ptr {
		rangeable = rangeable.Elem()
	}
	switch rangeable.Kind() {
	case reflect.Map:
		return i.mapFlow.Take(vm)
	case reflect.Slice, reflect.Array:
		return i.sliceOrArrayFlow.Take(vm)
	case reflect.Int:
		return i.intFlow.Take(vm)
	case reflect.String:
		// iterate over runes
		str := rangeable.String()
		runeSlice := make([]rune, 0, len(str))
		for _, r := range str {
			runeSlice = append(runeSlice, r)
		}
		vm.pushOperand(reflect.ValueOf(runeSlice))
		return i.sliceOrArrayFlow.Take(vm)
	case reflect.Func:
		// TODO
		// func(yield func(V) bool): //iter.Seq[V any]:
		//return i.funcFlow.Take(vm)
		fallthrough
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
