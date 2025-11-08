![gi logo](docs/gi-logo.png)

[![Go](https://github.com/emicklei/gi/actions/workflows/go.yml/badge.svg)](https://github.com/emicklei/gi/actions/workflows/go.yml)
[![GoDoc](https://pkg.go.dev/badge/github.com/emicklei/gi)](https://pkg.go.dev/github.com/emicklei/gi)
[![codecov](https://codecov.io/gh/emicklei/gi/branch/main/graph/badge.svg)](https://codecov.io/gh/emicklei/gi)
[![examples](https://img.shields.io/badge/dynamic/json?url=https://ernestmicklei.com/treerunner-report.json&query=$.label&label=gobyexample)]

a Go interpreter that can be used in plugins and debuggers.

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
- `GI_DOT=out.dot` : produce a Graphviz DOT file showing the call graph.

## Use as package

### run a program

```go
package main

import "github.com/emicklei/gi"

func main() {
	pkg, _ := gi.ParseSource(`package main

import "fmt"

func Hello(name string) {
	fmt.Println("Hello,", name)
}
`)
	gi.Call(pkg, "Hello", "3i/Atlas")
}
```

&copy; 2025. https://ernestmicklei.com . MIT License