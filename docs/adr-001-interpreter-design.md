# ADR-001: Go Interpreter Design

## Status

Proposed

## Context

The `gi` project requires a lightweight, embeddable Go interpreter for use in plugins and debuggers. This ADR outlines the architectural decisions behind the interpreter's design, focusing on how it implements key features of the Go specification. The target audience is advanced Go developers who need to understand the interpreter's internals for extension or maintenance.

## Decision

The interpreter is designed as a stack-based virtual machine (VM) that executes a custom Abstract Syntax Tree (AST). This approach was chosen for its simplicity, portability, and ease of implementation. The core components are:

- **AST Builder (`internal/ast_builder.go`)**: Go source code is parsed into a custom AST using a visitor pattern on the standard `go/ast` package. This custom AST is tailored for direct execution by the VM, simplifying the evaluation logic.

- **Stack-Based Virtual Machine (`internal/vm.go`)**: The VM executes the custom AST, using a stack to manage operands and function call frames. This design is efficient and straightforward to debug.

- **Built-in Functions (`internal/call_builtin.go`)**: Go's built-in functions (e.g., `len`, `append`, `make`) are implemented as special cases within the VM, directly manipulating the operand stack and memory.

### Implementation of Go Features

- **Types and Variables**: Go's type system is mapped to `reflect.Value` and `reflect.Type`, allowing the interpreter to handle Go's rich type system dynamically. Variables are managed in an environment (`internal/env.go`) that scopes them to function calls.

- **Control Flow**: `if`, `for`, and `switch` statements are implemented as nodes in the custom AST, with the VM controlling the execution flow based on the evaluation of conditions.

- **Functions**: Function calls create new frames on the VM's call stack, each with its own environment for local variables. This mirrors how Go manages function calls.

- **Pointers**: Pointers are implemented using `reflect.Value`'s pointer manipulation capabilities, allowing the interpreter to work with memory addresses.

## Consequences

### Pros

- **Simplicity**: The stack-based VM and custom AST are relatively easy to understand and maintain.
- **Portability**: The interpreter is written in pure Go and has no platform-specific dependencies.
- **Extensibility**: New language features can be added by extending the AST and adding corresponding evaluation logic to the VM.

### Cons

- **Performance**: The interpreter is significantly slower than compiled Go code due to the overhead of the VM and the use of `reflect`.
- **Incomplete Implementation**: Not all Go features are supported (e.g., goroutines, channels). A comprehensive list of supported features is in `STATUS.md`.
