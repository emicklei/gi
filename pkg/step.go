package pkg

import (
	"fmt"
	"go/token"
	"strconv"

	"github.com/emicklei/dot"
)

// step is a abstract type for linking steps in a control flow graph.
type step struct {
	id   int // set by graphBuilder
	next Step
	//pos  token.Pos
}

func (s *step) ID() int {
	return s.id
}

func (s *step) SetID(id int) {
	s.id = id
}

func (s *step) String() string {
	if s == nil {
		return "step(<nil>)"
	}
	return fmt.Sprintf("%2d: ?", s.id)
}

func (s *step) StringWith(label string) string {
	return fmt.Sprintf("%2d: %s", s.id, label)
}

func (s *step) Next() Step {
	return s.next
}

func (s *step) SetNext(n Step) {
	s.next = n
}

func (s *step) eval(vm *VM) {}

func (s *step) take(vm *VM) {
	vm.currentFrame.step = s.next
}

func (s *step) traverse(g *dot.Graph, fs *token.FileSet) dot.Node {
	return s.traverseWithLabel(g, s.String(), cursor(fs, s.pos()), fs)
}

func (s *step) traverseWithLabel(g *dot.Graph, label, edge string, fs *token.FileSet) dot.Node {
	sid := strconv.Itoa(s.ID())
	if n, ok := g.FindNodeById(sid); ok {
		return n
	}
	n := g.Node(sid).Label(label)
	if s.next != nil {
		nextN := s.next.traverse(g, fs)
		n.Edge(nextN, edge)
	}
	return n
}

func (s *step) pos() token.Pos {
	return token.NoPos
}

var _ Step = (*evaluableStep)(nil)

type evaluableStep struct {
	step
	Evaluable
}

func (s *evaluableStep) eval(vm *VM) {
	s.Evaluable.eval(vm)
}

func (s *evaluableStep) take(vm *VM) {
	s.Evaluable.eval(vm)
	if vm.currentFrame.step == s {
		// if the evaluable did not change the step, then move to the next step.
		vm.currentFrame.step = s.next
	}
}

func (s *evaluableStep) pos() token.Pos {
	return s.Evaluable.pos()
}

func (s *evaluableStep) String() string {
	if s == nil {
		return "evaluableStep(<nil>)"
	}
	return fmt.Sprintf("%d: %v", s.id, s.Evaluable)
}

func (s *evaluableStep) traverse(g *dot.Graph, fs *token.FileSet) dot.Node {
	return s.traverseWithLabel(g, s.String(), cursor(fs, s.pos()), fs)
}

type conditionalStep struct {
	step
	conditionFlow Step
	elseFlow      Step
}

func (c *conditionalStep) traverse(g *dot.Graph, fs *token.FileSet) dot.Node {
	if c.conditionFlow != nil {
		c.conditionFlow.traverse(g, fs)
	}
	me := c.step.traverseWithLabel(g, c.String(), "true", fs)
	if c.elseFlow != nil {
		sid := strconv.Itoa(c.elseFlow.ID())
		// no edge if visited before
		_, ok := g.FindNodeById(sid)
		if !ok {
			falseN := c.elseFlow.traverse(g, fs)
			me.Edge(falseN, "false")
		}
	}
	return me
}

func (c *conditionalStep) take(vm *VM) {
	if c.conditionFlow == nil {
		vm.currentFrame.step = c.next
		return
	}
	cond := vm.popOperand()
	if cond.Bool() {
		vm.currentFrame.step = c.next
		return
	}
	vm.currentFrame.step = c.elseFlow
}

func (c *conditionalStep) pos() token.Pos {
	if c.conditionFlow != nil {
		return c.conditionFlow.pos()
	}
	if c.elseFlow != nil {
		return c.elseFlow.pos()
	}
	return c.step.pos()
}

func (c *conditionalStep) String() string {
	if c == nil {
		return "conditionalStep(<nil>)"
	}
	return fmt.Sprintf("%d: ~if", c.ID())
}

type pushEnvironmentStep struct {
	step
	stmtPos token.Pos
}

func (p *pushEnvironmentStep) pos() token.Pos {
	return p.stmtPos
}

// TODO replace with funcStep
func newPushEnvironmentStep(pos token.Pos) *pushEnvironmentStep {
	return &pushEnvironmentStep{stmtPos: pos}
}

func (p *pushEnvironmentStep) take(vm *VM) {
	vm.currentFrame.pushEnv()
	vm.currentFrame.step = p.next
}

func (p *pushEnvironmentStep) String() string {
	if p == nil {
		return "pushEnvironmentStep(<nil>)"
	}
	return fmt.Sprintf("%d: ~push env", p.ID())
}
func (p *pushEnvironmentStep) traverse(g *dot.Graph, fs *token.FileSet) dot.Node {
	return p.step.traverseWithLabel(g, p.String(), cursor(fs, p.pos()), fs)
}

// TODO replace with funcStep
type popEnvironmentStep struct {
	step
	stmtPos token.Pos
}

func (p *popEnvironmentStep) pos() token.Pos {
	return p.stmtPos
}

func (p *popEnvironmentStep) take(vm *VM) {
	vm.currentFrame.popEnv()
	vm.currentFrame.step = p.next
}

func (p *popEnvironmentStep) String() string {
	if p == nil {
		return "popEnvironmentStep(<nil>)"
	}
	return fmt.Sprintf("%d: ~pop env", p.ID())
}

func (p *popEnvironmentStep) traverse(g *dot.Graph, fs *token.FileSet) dot.Node {
	return p.step.traverseWithLabel(g, p.String(), cursor(fs, p.pos()), fs)
}

type labeledStep struct {
	step
	label   string
	stmtPos token.Pos
}

// newLabeledStep creates a labeled step but does not add it to the current flow.
func newLabeledStep(label string, pos token.Pos) *labeledStep {
	return &labeledStep{label: label, stmtPos: pos}
}

func (s *labeledStep) pos() token.Pos {
	return s.stmtPos
}

func (s *labeledStep) SetPos(update token.Pos) {
	s.stmtPos = update
}

func (s *labeledStep) String() string {
	if s == nil {
		return "labeledStep(<nil>)"
	}
	return fmt.Sprintf("%2d: %v", s.id, s.label)
}

func (s *labeledStep) traverse(g *dot.Graph, fs *token.FileSet) dot.Node {
	return s.step.traverseWithLabel(g, s.String(), cursor(fs, s.pos()), fs)
}

type popOperandStep struct {
	step
	stmtPos token.Pos
}

func (p popOperandStep) pos() token.Pos {
	return p.stmtPos
}

func (p popOperandStep) take(vm *VM) {
	vm.popOperand()
	vm.currentFrame.step = p.next
}

func (p popOperandStep) String() string {
	return fmt.Sprintf("%d: ~pop operand", p.ID())
}

func (p popOperandStep) traverse(g *dot.Graph, fs *token.FileSet) dot.Node {
	return g.Node(strconv.Itoa(p.ID())).Label(p.String())
}

var _ Step = (*funcStep)(nil)

type funcStep struct {
	step
	stmtPos token.Pos
	label   string
	fun     func(vm *VM)
}

// newFuncStep creates a new funcStep with the given position and function to execute.
func newFuncStep(pos token.Pos, label string, fun func(vm *VM)) *funcStep {
	return &funcStep{stmtPos: pos, label: label, fun: fun}
}

func (p *funcStep) take(vm *VM) {
	p.fun(vm)
	if vm.currentFrame.step == p {
		// if the function did not change the step, then move to the next step.
		vm.currentFrame.step = p.next
	}
}

func (e funcStep) pos() token.Pos {
	return e.stmtPos
}

func (e funcStep) String() string {
	return fmt.Sprintf("%d: ~exec %s", e.ID(), e.label)
}

func (e funcStep) traverse(g *dot.Graph, fs *token.FileSet) dot.Node {
	return e.step.traverseWithLabel(g, fmt.Sprintf("%2d: ~exec %s", e.ID(), e.label), cursor(fs, e.pos()), fs)
}
