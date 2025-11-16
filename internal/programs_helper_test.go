package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"reflect"
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
					fmt.Fprintf(vm.output, "%v", rv.Elem().Interface())
				} else {
					fmt.Fprintf(vm.output, "%v", rv.Interface())
				}
			} else {
				if s, ok := a.(string); ok {
					io.WriteString(vm.output, s)
					continue
				} else if a == undeclaredNil {
					fmt.Fprintf(vm.output, "<nil>")
				} else if a == untypedNil {
					fmt.Fprintf(vm.output, "<nil>")
				} else {
					fmt.Fprintf(vm.output, "%v", a)
				}
			}
		}
	}))
}

func parseForDebug(t *testing.T, source string) *VM {
	t.Helper()
	if trace {
		fmt.Println("DEBUG:", t.Name())
	}
	pkg := buildPackage(t, source)
	vm := newVM(pkg.Env)
	vm.setFileSet(pkg.Fset)
	return vm
}

func parseAndWalk(t *testing.T, source string) string {
	t.Helper()
	if trace {
		fmt.Println("TEST:", t.Name())
	}
	pkg := buildPackage(t, source)
	vm := newVM(pkg.Env)
	vm.setFileSet(pkg.Fset)
	collectPrintOutput(vm)

	if trace {
		// create dot graph for debugging
		os.WriteFile(fmt.Sprintf("testgraphs/%s.src", t.Name()), []byte(source), 0644)
		gidot := fmt.Sprintf("testgraphs/%s.dot", t.Name())
		pkg.writeCallGraph(gidot)
		// will fail in pipeline without graphviz installed
		exec.Command("dot", "-Tsvg", "-o", gidot+".svg", gidot).Run()

		// create ast dump for debugging
		pkg.writeAST(fmt.Sprintf("testgraphs/%s.ast", t.Name()))
	}
	if _, err := CallPackageFunction(pkg, "main", nil, vm); err != nil {
		t.Fatal(err)
	}
	return vm.output.String()
}

func testProgramIn(t *testing.T, dir string, wantFuncOrString any) {
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
