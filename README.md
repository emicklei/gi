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

    gi run .

For development, the following environment variables control the execution and output:

- `GI_TRACE=1` : produce tracing of the virtual machine that executes the statements and expressions.
- `GI_CALL=out.dot` : produce a Graphviz DOT file showing the call graph.
- `GI_AST=out.ast` : produce the mirror AST text file.

## Use as package

### run a program

```go
package main

import "github.com/emicklei/gi"

func main() {
	pkg, _ := gi.ParseSource(`package main

import "fmt"

func Hello(name string) int {
	fmt.Println("Hello,", name)
	return 42
}
`)
	answer, err := gi.Call(pkg, "Hello", "3i/Atlas")
}
```

#### Credits

The build pipeline uses all programs of [Go By Example](https://gobyexample.com/) to check whether they are executable with `gi`.

&copy; 2025. https://ernestmicklei.com . MIT License