package internal

import (
	"bytes"
	"go/token"
	"os"
	"path"
	"testing"

	"golang.org/x/tools/go/packages"
)

func BenchmarkIfElseIfElse(b *testing.B) {
	src := `package main

func main() {
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			if i == j {
				print("a")
			} else if i < j {
				print("b")
			} else {
				print("c")
			}
		}
	}
}`
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedFiles,
		Fset: token.NewFileSet(),
		Dir:  path.Join(cwd, "../examples"),
		Overlay: map[string][]byte{
			path.Join(cwd, "../examples/main.go"): []byte(src),
		},
	}
	pkg, err := LoadPackage(cfg.Dir, cfg)
	if err != nil {
		b.Fatalf("failed to load packages: %v", err)
	}
	{
		b.Run("native", func(b *testing.B) {
			buf := new(bytes.Buffer)
			for i := 0; i < 100; i++ {
				for j := 0; j < 100; j++ {
					if i == j {
						buf.WriteString("a")
					} else if i < j {
						buf.WriteString("b")
					} else {
						buf.WriteString("c")
					}
				}
			}
		})
	}
	{
		prog, err := BuildPackage(pkg, false)
		if err != nil {
			b.Fatalf("failed to build program: %v", err)
		}
		b.Run("run", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				vm := newVM(prog.Env)
				collectPrintOutput(vm)
				if _, err := RunPackageFunction(prog, "main", nil, vm); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
	{
		prog, err := BuildPackage(pkg, true)
		if err != nil {
			b.Fatalf("failed to build program: %v", err)
		}
		b.Run("walk", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				vm := newVM(prog.Env)
				collectPrintOutput(vm)
				if err := WalkPackageFunction(prog, "main", vm); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
