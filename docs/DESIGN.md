# Gi Design

This document outlines the software design and architecture of the Gi Go interpreter.

## Overview

Gi is designed to interpret a Go program by parsing its source code into an Abstract Syntax Tree (AST) and then walking this tree to execute the program's logic. The primary goal is to enable interactive debugging and a REPL-like experience for Go. A secondary goal is to enable plugins in Go to dynamically extend Go programs.

The interpreter is composed of several key components that work together to manage the program's state, execution flow, and interaction with Go's language features.

## Core Components

### 1. Parsing

- **Input:** A valid Go program, including its sub-packages.
- **Process:** Gi uses the standard `go/parser` package to parse all the `.go` source files into an AST. This AST provides a structured, tree-like representation of the source code that is ideal for interpretation. The AST is transformed into a mirror AST which nodes have the logic for execution.
- **Output:** A complete mirror AST for the entire program.

### 2. Program Representation

- **Builder:** The `internal/step_builder.go` component takes the source AST and constructs an internal representation of the program. This involves creating the necessary data structures for execution and a call graph using `internal/graph_builder.go`.

### 3. The Virtual Machine (VM)

- **Engine:** The core of the interpreter is the Virtual Machine (VM), located in `internal/vm.go`. The VM is responsible for driving the execution of the program.
- **Execution Loop:** It operates on a step-by-step basis, processing one statement or expression at a time. This design is fundamental to allowing features like step-through debugging.

### 4. State Management

- **Environment (`env.go`):** This component manages the state of variables, constants, and functions within different scopes. It handles variable declarations, assignments, and lookups.
- **Call Stack (`stack.go`):** The call stack tracks active function calls. Each time a function is invoked, a new frame is pushed onto the stack, containing the function's local variables and execution context. When a function returns, its frame is popped.

## Execution Flow

The interpretation process for various language constructs is handled by dedicated components:

- **Statements (`stmt.go`):** A central dispatcher for handling different types of statements (e.g., `if`, `for`, `return`).
- **Expressions:** Specific files are dedicated to evaluating different kinds of expressions:
  - `binaryexpr.go`: Handles binary operations (`+`, `-`, `*`, `/`, `==`, etc.).
  - `unary.go`: Handles unary operations (`-`, `!`).
  - `call.go`: Manages function calls, including pushing to the call stack and passing arguments.
  - `literal.go`: Processes literal values (integers, strings, etc.).
- **Control Flow:**
  - `if.go`: Implements the logic for `if-else` statements.
  - `return.go`: Handles `return` statements, passing return values back to the caller.
- **Assignments (`assign.go`):** Manages the assignment of values to variables.

## External Code and Built-ins

- **Standard Library & External Packages:** Code from packages outside the main program (standard library, `go.mod` dependencies) is not interpreted directly. Instead, Gi uses Go's `reflect` package to interact with these compiled packages, allowing the interpreted program to call their functions and use their types. The `internal/stdlib_generated.go` file likely contains pre-generated wrappers for standard library functions.
- **Built-in Functions (`builtins.go`):** Go's built-in functions (e.g., `len`, `cap`, `append`) are implemented natively within the interpreter to ensure correct and efficient behavior.


## Why Flow returns head

Flow() builds a callgraph for a given Mirror Node (replicated Go AST Node).
To use that callgraph, one needs to have a reference to the first step of the chain of steps.
The first step is the head of the chain.

### Dev Notes

- About types: https://github.com/golang/example/tree/master/gotypes
- stackframe on Go stack not heap?
- think about a driver api to do stepping,breakpoints
- how to handle concurrency. (eval -> native, walk -> simulated?)
- a literal eval is pushing the value as operand; other literals should do the same.
- fallthrough cannot be used with a type switch.
- declTable must be declSlice; the order is important , see TestDeclarationExample
- clear with a pointer to a var?
- drop Instance?, use https://stackoverflow.com/questions/57567466/create-a-struct-by-reflection-in-go  this only support Exported fields. Cannot use it.
- can each Flowable be a step instead decorated by a step?
- look like ZeroValue can be dropped
-  make the stats of gobyexample available as a badge in the project with a link to the lastest build step
- https://img.shields.io/badge/dynamic/json?url=https://github.com/badges/shields/raw/master/package.json&query=$.name&label=piet
- binaryexpr can be optimized by inspecting type on build time and cache Go function for the expression evaluation
- same can be done for unary expr?
. fmt.Println for Instances needs rework
- symbolstable and typestable can be merged into one
- github.com/fatih/structtag replace with some SDK pkg?
- how to handle returnsType of FuncType? and what if FuncType is using local pkg types?
- handle omitzero


## potential blockers
- reflect structs can only have exposed fields. for that reason Instance was created but the SDK is not aware of this. For example, fmt.Println might not work correctly with Instance values.
- stepping happens per go-routine; what to do with the others when controlling one of them?

## external pkg
- if a program imports external packages then a new `gi` is created using additional generated sources that will setup all exported functions,consts and vars to the environment. this technique is also applied in `varvoy`.

## ideas for Gi Playground
- https://godbolt.org/
- call graphs in tab view
- structexplorer to see all environments and stackframes
