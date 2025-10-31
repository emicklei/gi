![gi logo](docs/gi-logo.png)

[![Go](https://github.com/emicklei/gi/actions/workflows/go.yml/badge.svg)](https://github.com/emicklei/gi/actions/workflows/go.yml)
[![GoDoc](https://pkg.go.dev/badge/github.com/emicklei/gi)](https://pkg.go.dev/github.com/emicklei/gi)
[![codecov](https://codecov.io/gh/emicklei/gi/branch/main/graph/badge.svg)](https://codecov.io/gh/emicklei/gi)

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
- `GI_STEP=1` : use the call graph of steps to execute the program; use the mirror AST otherwise.
- `GI_DOT=out.dot` : produce a Graphviz DOT file showing the call graph.

## Use as package

### run a program

```go
package main

import (
    "github.com/emicklei/gi"
)

func main() {
    gi.Run("path/to/main.go") // or gi.Run(".")       
}
```

&copy; 2025. https://ernestmicklei.com . MIT License