# Gi Design

This document outlines the software design and architecture of the Gi Go interpreter.

## Overview

Gi is designed to interpret a Go program by parsing its source code into an Abstract Syntax Tree (AST) and then executing it step-by-step. The interpreter maintains the program's state, including variable values, function calls, and control flow, to allow for features like debugging and interactive execution.

## Core Components

### 1. Parsing

- **Input:** A valid Go program, including its sub-packages.
- **Process:** Gi uses the standard `go/parser` package to parse all the `.go` source files into an AST. This AST provides a structured, tree-like representation of the source code that is ideal for interpretation. The AST is transformed into a mirror AST which nodes have the logic for execution.
- **Output:** A complete mirror AST for the entire program.

### 2. Program Representation

- **Builder:** The `pkg/graph_builder.go` component takes the mirror AST and constructs call graphs. These graphs represent the flow of function calls and control structures in the program, allowing for efficient execution and debugging.

### 3. The Virtual Machine (VM)

- **Engine:** The core of the interpreter is the Virtual Machine (VM), located in `pkg/vm.go`. The VM is responsible for driving the execution of the program.
- **Execution Loop:** It operates on a step-by-step basis, processing one statement or expression at a time using the `take` method. This design is fundamental to allowing features like step-through debugging.

### 4. State Management

- **Environment (`env.go`):** This component manages the state of variables, constants, and functions within different scopes. It handles variable declarations, assignments, and lookups. There is a hierarchical structure of environments, with the root being a `PackageEnvironment`. Each function call creates a new environment that can access its parent environment.
- **Call Stack (`stack.go`):** The call stack tracks active function calls. Each time a function is invoked, a new frame is pushed onto the stack, containing the function's local variables and execution context (which step to take). When a function returns, its frame is popped and results are pushed on the operands stack of the calling stack frame.

## Execution Flow

- **Call graphs:** The call graphs built by the builder are used to determine the order of execution. When a function is called, the VM consults the current step (`currentFrame.step`) to determine which step to execute. Nodes from the mirror AST participate in the execution can be part of the call graph and need to implement the `Evaluable` interface. After executing a step, it updates the current frame and the VM continues to the next step until the program finishes or a breakpoint is hit.

## External Code and Built-ins

- **Standard Library & External Packages:** Code from packages outside the main program (standard library, `go.mod` dependencies) is not interpreted directly. Instead, Gi uses Go's `reflect` package to interact with these compiled packages, allowing the interpreted program to call their functions and use their types. The `pkg/stdlib_generated.go` file likely contains pre-generated wrappers for standard library functions.
- **Built-in Functions (`builtins.go`):** Go's built-in functions (e.g., `len`, `cap`, `append`) are implemented natively within the interpreter to ensure correct and efficient behavior.

## Detailed Designs per Language Feature

- defers
- const declarations
- iota
- type declarations
- struct types and struct literals
- interface types and interface literals
- function declarations and function literals
- range loops

### defer

Both function declarations and function literals must handle deferred statements. Because the call of such a statement must capture any function arguments before executing it, a `funcInvocation` is created and added to the list of `defers` field in the stackframe created for a `*FuncDecl` or a `*FuncLit` at runtime.

#### panic recovery

When a panic occurs, the VM will execute all deferred statements in the current stack frame in LIFO order. If any of these deferred statements themselves cause a panic, the process continues until all deferred statements have been executed. After that, the original panic is re-raised.

### why `flow` returns head

`flow(*graphBuilder)` builds a callgraph for a given Mirror Node (replicated Go AST Node).
To use that callgraph, one needs to have a reference to the first step of the chain of steps.
The first step is the head of the chain.

### not all idents are equal

- new(int)
- int(0)
- var i int
- case int: 
- j := i.(int)




## operand order on stack
- when a function returns 2 or more values then operands are stacked in reverse order `pushCallResults` ; first is on top.
- for a multi-assign stmt the values of the rhs expressions must be stacked in reverse order:
   
  a , b := callReturning12() // stack: 2,1
  c , d := 1 , 2 // stack: 2,1

- so in assign, pairwise flow must be done right-to-left, eval must be done left-to-right

  d = 2, c = 1


# debugging 

- https://github.com/dbaumgarten/yodk/blob/v0.1.13/pkg/debug/handler.go#L23
- https://github.com/xhd2015/dlv-mcp

## structure building phases

- parsing Go source
- building mirror AST
- building flow graphs

## types
- builtintypes: int, *int
- SDK types: time.Time, sync.WaitGroup
- Interpreted struct types: type Aircraft struct{}
- Interpreted embedding struct types: type Aircraft struct{ Asset }
- Interpreted extended builtin types: type Count int
- Interpreted extended struct types: type Plane Aircraft
