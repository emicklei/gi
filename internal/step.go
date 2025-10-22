package internal

import (
	"fmt"
	"strconv"

	"github.com/emicklei/dot"
)

var _ Step = (*step)(nil)

var idgen int = 0

type step struct {
	id   int
	next Step
	Evaluable
}

func newStep(e Evaluable) *step {
	idgen++
	return &step{id: idgen, Evaluable: e}
}

type conditionalStep struct {
	*step
	conditionFlow Step
	elseFlow      Step
}

func (c *conditionalStep) Traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node {
	c.conditionFlow.Traverse(g, visited)
	me := c.step.traverse(g, c.step.String(), "true", visited)
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
	cond := vm.callStack.top().pop()
	if cond.Bool() {
		return c.next
	}
	return c.elseFlow
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

func (s *step) ID() int {
	return s.id
}

func (s *step) String() string {
	if s == nil {
		return "nil"
	}
	return fmt.Sprintf("%2d:step(%v)", s.id, s.Evaluable)
}

func (s *step) Next() Step {
	return s.next
}

func (s *step) SetNext(n Step) {
	s.next = n
}

func (s *step) Take(vm *VM) Step {
	if s.Evaluable != nil {
		s.Evaluable.Eval(vm)
	}
	return s.next
}

type pushStackFrameStep struct {
	*step
}

func (p *pushStackFrameStep) String() string { return fmt.Sprintf("%2d:step(push stackframe)", p.ID()) }

func (p *pushStackFrameStep) Traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node {
	return p.step.traverse(g, p.String(), "next", visited)
}

func (p *pushStackFrameStep) Take(vm *VM) Step {
	vm.pushNewFrame(p.Evaluable)
	return p.next
}

type popStackFrameStep struct {
	*step
}

func (p *popStackFrameStep) Take(vm *VM) Step {
	vm.popFrame()
	return p.next
}

func (p *popStackFrameStep) String() string { return fmt.Sprintf("%2d:step(pop stackframe)", p.ID()) }

func (p *popStackFrameStep) Traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node {
	return p.step.traverse(g, p.String(), "next", visited)
}
