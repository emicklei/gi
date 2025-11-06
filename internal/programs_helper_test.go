package internal

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path"
	"reflect"
	"testing"

	"golang.org/x/tools/go/packages"
)

func buildPackage(t *testing.T, source string) *Package {
	t.Helper()
	cwd, _ := os.Getwd()
	cfg := &packages.Config{
		// copied from Package.go
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedFiles | packages.NeedTypesInfo,
		Fset: token.NewFileSet(),
		Dir:  path.Join(cwd, "./testprogram"),
		// copied from Package.go
		ParseFile: func(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
			return parser.ParseFile(fset, filename, src, parser.SkipObjectResolution)
		},
		Overlay: map[string][]byte{
			path.Join(cwd, "./testprogram/main.go"): []byte(source),
		},
	}
	gopkg, err := LoadPackage(cfg.Dir, cfg)
	if err != nil {
		t.Fatalf("failed to load packages: %v", err)
	}
	ffpkg, err := BuildPackage(gopkg)
	if err != nil {
		t.Fatalf("failed to build package: %v", err)
	}
	return ffpkg
}

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
				} else {
					fmt.Fprintf(vm.output, "%v", a)
				}
			}
		}
	}))
}

func parseAndWalk(t *testing.T, source string) string {
	t.Helper()
	if trace {
		fmt.Println("test:", t.Name())
	}
	pkg := buildPackage(t, source)
	vm := newVM(pkg.Env)
	collectPrintOutput(vm)

	// create dot graph for debugging
	os.WriteFile(fmt.Sprintf("testgraphs/%s.src", t.Name()), []byte(source), 0644)
	gidot := fmt.Sprintf("testgraphs/%s.dot", t.Name())
	pkg.writeDotGraph(gidot)
	// will fail in pipeline without graphviz installed
	exec.Command("dot", "-Tsvg", "-o", gidot+".svg", gidot).Run()

	if _, err := CallPackageFunction(pkg, "main", nil, vm); err != nil {
		panic(err)
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
