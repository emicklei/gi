package pkg

import "github.com/google/go-dap"

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

func (a *DAPAccess) Launch(functionName string, args []any) {
	a.vm.Launch(functionName, args)
}
func (a *DAPAccess) Next() {
	a.vm.Next()
}
func (a *DAPAccess) Threads() []dap.Thread {
	return []dap.Thread{
		{
			Id:   1,
			Name: "main",
		},
	}
}
