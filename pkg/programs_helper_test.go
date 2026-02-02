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
	vm.localEnv().set("print", reflect.ValueOf(func(args ...any) {
		for _, a := range args {
			if rv, ok := a.(reflect.Value); ok && rv.IsValid() && rv.CanInterface() {
				// check for pointer to array
				if rv.Kind() == reflect.Pointer && rv.Elem().Kind() == reflect.Array {
					fmt.Fprintf(vm.output, "%v", rv.Elem().Interface())
				} else if rv.Kind() == reflect.Pointer {
					v := rv.Elem().Interface()
					fmt.Fprintf(vm.output, "%v", v)
				} else {
					v := rv.Interface()
					fmt.Fprintf(vm.output, "%v", v)
				}
			} else {
				if s, ok := a.(string); ok {
					io.WriteString(vm.output, s)
					continue
					// handle *StructValue specially because it implements Format
				} else if psv, ok := a.(*StructValue); ok {
					fmt.Fprintf(vm.output, "%p", psv)
				} else if a == undeclaredNil {
					fmt.Fprintf(vm.output, "(0x0,0x0)")
				} else if a == untypedNil {
					fmt.Fprintf(vm.output, "(0x0,0x0)")
				} else {
					fmt.Fprintf(vm.output, "%v", a)
				}
			}
		}
	}))
}

func parseAndWalk(t *testing.T, source string) string {
	t.Helper()
	defer func() {
		if trace {
			fmt.Println("TESTED:", t.Name())
		}
	}()
	pkg := buildPackage(t, source)
	vm := NewVM(pkg.env)
	vm.setFileSet(pkg.Fset)
	collectPrintOutput(vm)

	if trace {
		// create dot graph for debugging
		os.WriteFile(fmt.Sprintf("internal/testgraphs/%s.src", t.Name()), []byte(source), 0644)
		dotFileName := fmt.Sprintf("internal/testgraphs/%s.dot", t.Name())
		pkg.writeCallGraph(dotFileName)
		// will fail in pipeline without graphviz installed
		exec.Command("dot", "-Tsvg", "-o", dotFileName+".svg", dotFileName).Run()
		os.Remove(dotFileName)

		// create ast dump for debugging, requires test to set attribute(s)
		astFileName := fmt.Sprintf("internal/testgraphs/%s", t.Name())
		if getAttr(t, "ast") == "true" {
			pkg.writeAST(astFileName + ".ast")
		}
		if getAttr(t, "go.ast") == "true" {
			writeGoAST(astFileName+".go.ast", pkg.Package)
		}
	}
	if _, err := CallPackageFunction(pkg, "main", nil, vm); err != nil {
		t.Fatal(err)
	}
	return vm.output.String()
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
	_, err = CallPackageFunction(pkg, "main", nil, nil)
	if err != nil {
		t.Fatalf("failed to run package in %s: %v", loc, err)
	}
}

func testMain(t *testing.T, source string, wantFuncOrString any) {
	t.Parallel()
	t.Helper()
	out := parseAndWalk(t, source)
	if fn, ok := wantFuncOrString.(func(string) bool); ok {
		if !fn(out) {
			t.Errorf("got [%v] which does not match predicate", out)
		}
		return
	}
	want := wantFuncOrString.(string)
	if got, want := out, want; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
}
