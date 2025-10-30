package internal

import (
	"fmt"
	"strconv"

	"github.com/emicklei/dot"
)

var _ Step = (*step)(nil)

type step struct {
	id        int32 // set by graphBuilder
	next      Step
	Evaluable // can be nil for structural steps
}

type conditionalStep struct {
	*step
	conditionFlow Step
	elseFlow      Step
}

func (c *conditionalStep) String() string {
	return fmt.Sprintf("%2d: if", c.ID())
}

func (c *conditionalStep) Traverse(g *dot.Graph, visited map[int32]dot.Node) dot.Node {
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
	cond := vm.frameStack.top().pop()
	if cond.Bool() {
		return c.next
	}
	return c.elseFlow
}

func (s *step) Traverse(g *dot.Graph, visited map[int32]dot.Node) dot.Node {
	return s.traverse(g, s.String(), "next", visited)
}

func (s *step) traverse(g *dot.Graph, label, edge string, visited map[int32]dot.Node) dot.Node {
	if n, ok := visited[s.id]; ok {
		return n
	}
	n := g.Node(strconv.FormatInt(int64(s.ID()), 10)).Label(label)
	visited[s.id] = n
	if s.next != nil {
		nextN := s.next.Traverse(g, visited)
		n.Edge(nextN, edge)
	}
	return n
}

func (s *step) ID() int32 {
	return s.id
}

func (s *step) String() string {
	if s == nil {
		return "nil"
	}
	return fmt.Sprintf("%2d: %v", s.id, s.Evaluable)
}

func (s *step) StringWith(label string) string {
	if s == nil {
		return "nil"
	}
	return fmt.Sprintf("%2d: %s", s.id, label)
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

func (p *pushStackFrameStep) String() string { return fmt.Sprintf("%2d: ~push stackframe", p.ID()) }

func (p *pushStackFrameStep) Traverse(g *dot.Graph, visited map[int32]dot.Node) dot.Node {
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

func (p *popStackFrameStep) String() string { return fmt.Sprintf("%2d: ~pop stackframe", p.ID()) }

func (p *popStackFrameStep) Traverse(g *dot.Graph, visited map[int32]dot.Node) dot.Node {
	return p.step.traverse(g, p.String(), "next", visited)
}

type returnStep struct {
	*step
}

func (r *returnStep) Traverse(g *dot.Graph, visited map[int32]dot.Node) dot.Node {
	return g.Node(strconv.FormatInt(int64(r.ID()), 10)).Label(r.String())
}

type labeledStep struct {
	*step
	label string
}

func (s *labeledStep) String() string {
	return fmt.Sprintf("%2d: %v", s.id, s.label)
}
func (s *labeledStep) Take(vm *VM) Step {
	return s.next
}
func (s *labeledStep) Traverse(g *dot.Graph, visited map[int32]dot.Node) dot.Node {
	return s.step.traverse(g, s.String(), "next", visited)
}
