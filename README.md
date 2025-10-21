# gi

[![Go](https://github.com/emicklei/gi/actions/workflows/go.yml/badge.svg)](https://github.com/emicklei/gi/actions/workflows/go.yml)
[![GoDoc](https://pkg.go.dev/badge/github.com/emicklei/gi)](https://pkg.go.dev/github.com/emicklei/gi)
[![codecov](https://codecov.io/gh/emicklei/gi/branch/main/graph/badge.svg)](https://codecov.io/gh/emicklei/gi)

a Go interpreter that can be used in plugins and debuggers.

![gi logo](docs/gi-logo.png)

## status

This is work in progress.
See [examples](./examples) for runnable examples using the `gi` cli.
See [status](STATUS.md) for the supported Go language features.

## install

    go install github.com/emicklei/gi/cmd/gi@latest

## Use CLI

    gi run .

## Use as package

```go
package main

import (
    "github.com/emicklei/gi"
)

func main() {
    gi.Run("path/to/program")        
}
```

&copy; 2025. https://ernestmicklei.com . MIT License