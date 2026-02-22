package pkg

import (
	"cmp"
	"go/token"
	"path/filepath"
	"slices"

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
func (a *DAPAccess) StackFrames(dap.StackTraceArguments) (frames []dap.StackFrame) {
	for _, eachFrame := range a.vm.callStack {
		var tokloc token.Position
		if eachFrame.callee != nil {
			tokloc = a.vm.pkg.Fset.Position(eachFrame.callee.pos())
		}
		dapFrame := dap.StackFrame{
			Id:   eachFrame.id,
			Name: stringOf(eachFrame.callee), // TODO
			Source: &dap.Source{
				Name: filepath.Base(tokloc.Filename),
				Path: tokloc.Filename,
			},
			Line:   tokloc.Line,
			Column: tokloc.Column,
		}
		frames = append(frames, dapFrame)
	}
	return
}

// Scopes describes the variable scopes that are available for the current stack frame.
func (a *DAPAccess) Scopes(dap.ScopesArguments) (scopes []dap.Scope) {
	here := a.vm.currentFrame.env
	for {
		if here == nil {
			break
		}
		scopes = here.appendScopes(scopes)
		here = here.parent()
	}
	// sort by scope name
	slices.SortFunc(scopes, func(s1, s2 dap.Scope) int { return cmp.Compare(s1.Name, s2.Name) })
	return
}

// Variables lists the variables for the provided scope reference.
func (a *DAPAccess) Variables(args dap.VariablesArguments) (vars []dap.Variable) {
	here := a.vm.currentFrame.env
	for {
		if here == nil {
			break
		}
		if args.VariablesReference == here.depth() {
			vars = here.appendVariables(vars)
			break
		}
		here = here.parent()
	}
	// sort by var name
	slices.SortFunc(vars, func(s1, s2 dap.Variable) int { return cmp.Compare(s1.Name, s2.Name) })
	return
}
