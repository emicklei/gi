package pkg

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func buildPackage(t *testing.T, source string) *Package {
	t.Helper()

	goPkg, err := ParseSource(source)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}
	// TODO until generics is done
	d := newGenericsDetector(goPkg)
	for _, stx := range goPkg.Syntax {
		for _, decl := range stx.Decls {
			d.Visit(decl)
		}
	}
	// END TODO

	giPkg, err := BuildPackage(goPkg)
	if err != nil {
		t.Fatalf("failed to build package: %v", err)
	}
	return giPkg
}

var stdfuncsMutex sync.Mutex

// this print function outputs are different from the standard and is only used for tests
func collectPrintOutput(vm *VM) {
	vm.pkg.env.valueSet("print", reflect.ValueOf(func(args ...any) {
		for _, a := range args {
			fmt.Fprint(vm.output, stringOf(a))
		}
	}))
	// maybe move this TODO
	stdfuncsMutex.Lock()
	defer stdfuncsMutex.Unlock()

	stdfuncs["fmt"]["Print"] = reflect.ValueOf(func(args ...any) {
		fmt.Fprint(vm.output, args...)
	})
	stdfuncs["fmt"]["Printf"] = reflect.ValueOf(func(args ...any) {
		fmt.Fprintf(vm.output, args[0].(string), args[1:]...)
	})
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

func debug(t *testing.T) func() {
	setAttr(t, "dot", 1)
	trace = true
	return func() {
		trace = false
	}
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
	vm := NewVM(pkg)
	collectPrintOutput(vm)

	if getAttr(t, "dot") != nil {
		// create dot graph for debugging
		os.WriteFile(fmt.Sprintf("internal/testgraphs/%s.src", t.Name()), []byte(source), 0644)
		dotFileName := fmt.Sprintf("internal/testgraphs/%s.dot", t.Name())
		pkg.writeCallGraph(dotFileName)
		// will fail in pipeline without graphviz installed
		exec.Command("dot", "-Tsvg", "-o", dotFileName+".svg", dotFileName).Run()
		os.Remove(dotFileName)
	}
	vm.launch("main", nil)
	for {
		if err := vm.Next(); err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}
	}
	// check output
	got := vm.output.String()
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

// used?
func isPointerToStructValue(v reflect.Value) bool {
	if v.Kind() != reflect.Pointer {
		return false
	}
	if v.Elem().Kind() != reflect.Struct {
		return false
	}
	if v.Elem().Type().Name() != "StructValue" {
		return false
	}
	// not exact package match to allow source code forks
	if !strings.HasSuffix(v.Elem().Type().PkgPath(), "/gi/pkg") {
		return false
	}
	return true
}

// used?
func isNonSDKFunction(rv reflect.Value) bool {
	if !rv.IsValid() {
		return false
	}
	if !rv.CanInterface() {
		return false
	}
	switch rv.Interface().(type) {
	case builtinFunc:
		return true
	case *FuncDecl:
		return true
	case *FuncLit:
		return true
	}
	return false
}
