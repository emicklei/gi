![gi logo](docs/gi-logo.png)

[![Go](https://github.com/emicklei/gi/actions/workflows/go.yml/badge.svg)](https://github.com/emicklei/gi/actions/workflows/go.yml)
[![GoDoc](https://pkg.go.dev/badge/github.com/emicklei/gi)](https://pkg.go.dev/github.com/emicklei/gi)
[![codecov](https://codecov.io/gh/emicklei/gi/branch/main/graph/badge.svg)](https://codecov.io/gh/emicklei/gi)
[![examples](https://img.shields.io/badge/dynamic/json?url=https://ernestmicklei.com/treerunner-report.json&query=$.label&label=gobyexample)](https://ernestmicklei.com/treerunner-report.json)

`gi` is a Go interpreter that creates an executable representation of a Go program from source.
It offers a virtual machine that can step through such a program and allows access to the full stack and scoped variables (environment).

`gi` is implemented using Go reflection API so one can expect 10x slower program execution depending on the complexity.

## mission

### runtime

- support latest Go SDK
- support type parameterization (generics)
- support Go modules (will require pre-compilation)

### debugging

- offer a DAP interface
- handle source changes during a debugging session:
	- change a function definition
	- change a struct definition
	- add package constant|variable
- debugging concurrent programs

## status

This is work in progress.
See [examples](./examples) for runnable examples using the `gi` cli.
See [status](STATUS.md) for the supported Go language features.

## install

    go install github.com/emicklei/gi/cmd/gi@latest

## Use CLI

### run

```bash
gi run .
```

### step

```bash
gi step .
```

## DAP server

```bash
gi dap --listen=127.0.0.1:52950 --log-dest=3 --log
```

For development, the following environment variables control the execution and output:

- `GI_TRACE=1` : produce tracing of the virtual machine that executes the statements and expressions.
- `GI_CALL=out.dot` : produce a Graphviz DOT file showing the call graph.

## Use as package

### run a program

```go
package main

import "github.com/emicklei/gi"

func main() {
	pkg, _ := gi.Parse(`package main

import "fmt"

func Hello(name string) int {
	fmt.Println("Hello,", name)
	return 42
}
`)
	answer, err := gi.Call(pkg, "Hello", "3i/Atlas")
}
```

### use of in-process DAP (Debug Adapter Protocol)

```go
	gopkg, _ := pkg.LoadPackage(".", nil)
	ipkg, _ := pkg.BuildPackage(gopkg)
	runner := pkg.NewDAPAccess(pkg.NewVM(ipkg))
	runner.Launch("main", nil)
	for {
		if err := runner.Next(); err != nil {
			if err == io.EOF {
				return
			}
		}
		_ = runner.Threads()
		_ = runner.StackFrames(...)
		_ = runner.Scopes(...)
		_ = runner.Variables(...)
	}
```

### Limitations

The following features are not supported:
- no functions/consts/vars from the following packages:
  - debug
  - go
  - syscall
  - runtime/testdata
- these symbols are not available because of NumericOverflow in reflect.Value:
  - hash/crc64.ECMA
  -	hash/crc64.ISO
  -	math.MaxUint
  -	math.MaxUint64

#### Credits

The build pipeline uses all programs of [Go By Example](https://gobyexample.com/) to check whether they are executable with `gi`.

&copy; 2026. https://ernestmicklei.com . MIT License
