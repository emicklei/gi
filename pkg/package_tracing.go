package pkg

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/emicklei/dot"
	"github.com/spewerspew/spew"
	"golang.org/x/tools/go/packages"
)

//
// for debugging
//

func (p *Package) writeAST(fileName string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("SPEW: failed to write AST file", r)
		}
	}()
	buf := new(bytes.Buffer)
	spew.Config.DisableMethods = true
	spew.Config.MaxDepth = 8 // TODO see if this is enough
	done := make(chan bool)
	go func() {
		// only dump the actual values of each var/function in the environment
		for _, v := range p.env.Env.(*Environment).valueTable {
			// skip SDKPackage
			val := v.Interface()
			if _, ok := val.(SDKPackage); ok {
				continue
			}
			spew.Fdump(buf, val)
		}
		done <- true
	}()
	select {
	case <-time.After(2 * time.Second):
		fmt.Println("AST writing took more than 2 seconds, aborting")
		close(done)
	case <-done:
	}
	os.WriteFile(fileName, buf.Bytes(), 0644)
}

func writeGoAST(fileName string, goPkg *packages.Package) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("SPEW: failed to write Go AST file", r)
		}
	}()
	buf := new(bytes.Buffer)
	// do not write fileset
	fs := goPkg.Fset
	goPkg.Fset = nil
	defer func() {
		goPkg.Fset = fs
	}()
	spew.Config.DisableMethods = true
	spew.Config.MaxDepth = 8 // TODO see if this is enough
	done := make(chan bool)
	go func() {
		spew.Fdump(buf, goPkg)
		done <- true
	}()
	select {
	case <-time.After(2 * time.Second):
		fmt.Println("Go AST writing took more than 2 seconds, aborting")
		close(done)
	case <-done:
	}
	os.WriteFile(fileName, buf.Bytes(), 0644)
}

func (p *Package) writeCallGraph(fileName string) {
	g := dot.NewGraph(dot.Directed)
	g.NodeInitializer(func(n dot.Node) {
		n.Box()
		n.Attr("fillcolor", "#EBFAFF") // https://htmlcolorcodes.com/
		n.Attr("style", "rounded,filled")
	})
	// setup flow
	if p.callGraph != nil {
		sub := g.Subgraph("pkg."+p.Name, dot.ClusterOption{})
		p.callGraph.traverse(sub, p.Fset)
	} else {
		fmt.Println("WARN: no call graph to write for package", p.Name)
	}

	// for each function in the package create a subgraph
	values := p.env.Env.(*Environment).valueTable
	for k, v := range values {
		if funDecl, ok := v.Interface().(*FuncDecl); ok {
			if funDecl.graph == nil {
				continue
			}
			sub := g.Subgraph(k, dot.ClusterOption{})
			funDecl.graph.traverse(sub, p.Fset)
		}
	}
	os.WriteFile(fileName, []byte(g.String()), 0644)
}
