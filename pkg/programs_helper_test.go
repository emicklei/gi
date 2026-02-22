package pkg

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"reflect"
	"sync"
	"testing"
)

func buildPackage(t *testing.T, source string) *Package {
	t.Helper()

	locakPkg, err := ParseSource(source)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}
	return locakPkg
}

// this print function outputs are different from the standard and is only used for tests
func collectPrintOutput(vm *VM) {
	vm.pkg.env.valueSet("print", reflect.ValueOf(func(args ...any) {
		for _, a := range args {
			fmt.Fprint(vm.output, stringOf(a))
		}
	}))
	// stdfuncs["fmt"]["Print"] = reflect.ValueOf(func(args ...any) {
	// 	for _, a := range args {
	// 		fmt.Fprint(vm.output, stringOf(a))
	// 	}
	// })
}

// Per-test attribute storage
var testAttrs = sync.Map{}

func setAttr(t *testing.T, key string, val any) {
	t.Helper()
	testAttrs.Store(fmt.Sprintf("%p.%s", t, key), val)
}
func getAttr(t *testing.T, key string) any {
	t.Helper()
	v, _ := testAttrs.Load(fmt.Sprintf("%p.%s", t, key))
	return v
}

func testProgramIn(t *testing.T, dir string, _ any) {
	// cannot be parallel because of os.Chdir
	t.Helper()
	cwd, _ := os.Getwd()
	loc := path.Join(cwd, dir)
	gopkg, err := LoadPackage(loc, nil)
	if err != nil {
		t.Fatalf("failed to load package in %s: %v", loc, err)
	}
	os.Chdir(loc)
	defer os.Chdir(cwd)
	pkg, err := BuildPackage(gopkg)
	if err != nil {
		t.Fatalf("failed to build package in %s: %v", loc, err)
	}
	_, err = NewVM(pkg).callPackageFunction("main", nil)
	if err != nil {
		t.Fatalf("failed to run package in %s: %v", loc, err)
	}
}

func testMain(t *testing.T, source string, wantFuncOrString any) {
	t.Parallel()
	t.Helper()
	defer func() {
		if trace {
			fmt.Println("TESTED:", t.Name())
		}
	}()
	pkg := buildPackage(t, source)
	runner := NewVM(pkg)
	collectPrintOutput(runner)

	if getAttr(t, "dot") != nil {
		// create dot graph for debugging
		os.WriteFile(fmt.Sprintf("internal/testgraphs/%s.src", t.Name()), []byte(source), 0644)
		dotFileName := fmt.Sprintf("internal/testgraphs/%s.dot", t.Name())
		pkg.writeCallGraph(dotFileName)
		// will fail in pipeline without graphviz installed
		exec.Command("dot", "-Tsvg", "-o", dotFileName+".svg", dotFileName).Run()
		os.Remove(dotFileName)
	}
	// create ast dump for debugging, requires test to set attribute(s)
	astFileName := fmt.Sprintf("internal/testgraphs/%s", t.Name())
	if getAttr(t, "ast") == "true" {
		pkg.writeAST(astFileName + ".ast")
	}
	if getAttr(t, "go.ast") == "true" {
		writeGoAST(astFileName+".go.ast", pkg.Package)
	}
	runner.launch("main", nil)
	// walk the steps of the program
	for {
		if err := runner.Next(); err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}
	}

	// check output
	got := runner.output.String()
	var want string
	switch v := wantFuncOrString.(type) {
	case string:
		want = v
	case func(string) bool:
		if !v(got) {
			t.Fatalf("output did not satisfy condition: %s", got)
		}
		return
	default:
		t.Fatalf("invalid want type: %T", wantFuncOrString)
	}
	if got != want {
		t.Fatalf("unexpected output: got %q, want %q", got, want)
	}
}
