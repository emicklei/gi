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

func (s *step) Eval(vm *VM) {}

func (s *step) Take(vm *VM) Step {
	return s.next
}

func (s *step) Traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node {
	return s.traverse(g, s.String(), "next", visited)
}

func (s *step) traverse(g *dot.Graph, label, edge string, visited map[int]dot.Node) dot.Node {
	if n, ok := visited[s.id]; ok {
		return n
	}
	n := g.Node(strconv.Itoa(s.ID())).Label(label)
	visited[s.id] = n
	if s.next != nil {
		nextN := s.next.Traverse(g, visited)
		n.Edge(nextN, edge)
	}
	return n
}

func (s *step) Pos() token.Pos {
	return token.NoPos
}

var _ Step = (*evaluableStep)(nil)

type evaluableStep struct {
	step
	Evaluable
}

func (s *evaluableStep) Eval(vm *VM) {
	s.Evaluable.Eval(vm)
}

func (s *evaluableStep) Take(vm *VM) Step {
	s.Evaluable.Eval(vm)
	return s.next
}

func (s *evaluableStep) Pos() token.Pos {
	return s.Evaluable.Pos()
}

func (s *evaluableStep) String() string {
	if s == nil {
		return "evaluableStep(<nil>)"
	}
	return fmt.Sprintf("%d: %v", s.id, s.Evaluable)
}

func (s *evaluableStep) Traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node {
	return s.traverse(g, s.String(), "next", visited)
}

type conditionalStep struct {
	step
	conditionFlow Step
	elseFlow      Step
}

func (c *conditionalStep) Traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node {
	c.conditionFlow.Traverse(g, visited)
	me := c.step.traverse(g, c.String(), "true", visited)
	if c.elseFlow != nil {
		// no edge if visited before
		if _, ok := visited[c.elseFlow.ID()]; ok {
			return me
		}
		falseN := c.elseFlow.Traverse(g, visited)
		me.Edge(falseN, "false")
	}
	return me
}

func (c *conditionalStep) Take(vm *VM) Step {
	if c.conditionFlow == nil {
		return c.next
	}
	cond := vm.popOperand()
	if cond.Bool() {
		return c.next
	}
	return c.elseFlow
}

func (c *conditionalStep) Pos() token.Pos {
	return c.conditionFlow.Pos()
}

func (c *conditionalStep) String() string {
	if c == nil {
		return "conditionalStep(<nil>)"
	}
	return fmt.Sprintf("%d: if", c.ID())
}

type pushEnvironmentStep struct {
	step
	pos token.Pos
}

func (p *pushEnvironmentStep) Pos() token.Pos {
	return p.pos
}

func newPushEnvironmentStep(pos token.Pos) *pushEnvironmentStep {
	return &pushEnvironmentStep{pos: pos}
}

func (p *pushEnvironmentStep) Take(vm *VM) Step {
	vm.currentFrame.pushEnv()
	return p.next
}

func (p *pushEnvironmentStep) String() string {
	if p == nil {
		return "pushEnvironmentStep(<nil>)"
	}
	return fmt.Sprintf("%d: ~push env", p.ID())
}
func (p *pushEnvironmentStep) Traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node {
	return p.step.traverse(g, p.String(), "next", visited)
}

type popEnvironmentStep struct {
	step
	pos token.Pos
}

func (p *popEnvironmentStep) Pos() token.Pos {
	return p.pos
}

func newPopEnvironmentStep(pos token.Pos) *popEnvironmentStep {
	return &popEnvironmentStep{pos: pos}
}

func (p *popEnvironmentStep) Take(vm *VM) Step {
	vm.currentFrame.popEnv()
	return p.next
}

func (p *popEnvironmentStep) String() string {
	if p == nil {
		return "popEnvironmentStep(<nil>)"
	}
	return fmt.Sprintf("%d: ~pop env", p.ID())
}

func (p *popEnvironmentStep) Traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node {
	return p.step.traverse(g, p.String(), "next", visited)
}

type returnStep struct {
	evaluableStep
}

func (r *returnStep) Traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node {
	return g.Node(strconv.Itoa(r.evaluableStep.ID())).Label(r.String())
}

type labeledStep struct {
	step
	label string
	pos   token.Pos
}

func (s *labeledStep) Pos() token.Pos {
	return s.pos
}

func (s *labeledStep) String() string {
	if s == nil {
		return "labeledStep(<nil>)"
	}
	return fmt.Sprintf("%2d: %v", s.id, s.label)
}

func (s *labeledStep) Traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node {
	return s.step.traverse(g, s.String(), "next", visited)
}

type popOperandStep struct {
	step
	pos token.Pos
}

func (p *popOperandStep) Pos() token.Pos {
	return p.pos
}

func (p *popOperandStep) Take(vm *VM) Step {
	vm.popOperand()
	return p.next
}

func (p *popOperandStep) String() string {
	return fmt.Sprintf("%d: ~pop operand", p.ID())
}
func (p *popOperandStep) Traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node {
	return g.Node(strconv.Itoa(p.ID())).Label(p.String())
}
