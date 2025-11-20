package internal

import (
	"testing"
)

func TestGrapherNext(t *testing.T) {
	g := newGraphBuilder(nil)
	e1 := Ident{}
	g.next(e1)
	s1 := g.current
	e2 := Ident{}
	g.next(e2)
	s2 := g.current

	if s1.Next() != s2 {
		t.Errorf("expected s1.Next() to be s2")
	}
	if g.previous != s1 {
		t.Errorf("expected g.previous to be s1")
	}
}
