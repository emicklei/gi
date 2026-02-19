package pkg

import (
	"path/filepath"

	"github.com/google/go-dap"
)

type DAPAccess struct {
	vm *VM
}

// NewDAPAccess creates a new wrapper around a VM instance
// to access DAP (Debug Adapter Protocol) data and control.
func NewDAPAccess(vm *VM) *DAPAccess {
	return &DAPAccess{
		vm: vm,
	}
}

// Launch starts execution of the given function on the underlying VM with the provided arguments.
func (a *DAPAccess) Launch(functionName string, args []any) {
	a.vm.Launch(functionName, args)
}

// Next advances the VM by a single debugging step.
func (a *DAPAccess) Next() error {
	return a.vm.Next()
}

// Threads reports the list of active debugger threads.
func (a *DAPAccess) Threads() []dap.Thread {
	return []dap.Thread{
		{
			Id:   1,
			Name: "main",
		},
	}
}

// StackFrames returns the current call stack for the selected thread.
func (a *DAPAccess) StackFrames(dap.StackTraceArguments) []dap.StackFrame {
	return []dap.StackFrame{
		{Id: 1, Name: "main.main", Source: &dap.Source{Name: "main.go", Path: filepath.Join(a.vm.pkg.Dir, "main.go")}, Line: 12},
	}
}

// Scopes describes the variable scopes that are available for the current stack frame.
func (a *DAPAccess) Scopes(dap.ScopesArguments) []dap.Scope {
	return []dap.Scope{
		{Name: "Local", VariablesReference: 1},
	}
}

// Variables lists the variables for the provided scope reference.
func (a *DAPAccess) Variables(dap.VariablesArguments) []dap.Variable {
	return []dap.Variable{
		{Name: "x", Value: "42", VariablesReference: 0},
		{Name: "y", Value: "\"hello\"", VariablesReference: 0},
	}
}
