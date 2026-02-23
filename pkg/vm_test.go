package pkg

import (
	"go/token"
	"io"
	"testing"
)

func TestVMNext(t *testing.T) {
	t.Skip()
	vm := NewVM(&Package{env: newPkgEnvironment(nil)})

	declare := newFuncStep(token.NoPos, "declare", func(vm *VM) {
		console("declare called")
	})

	target := newFuncStep(token.NoPos, "target", func(vm *VM) {
		console("target called")
	})

	initPkg := newFuncStep(token.NoPos, "initpkg", func(vm *VM) {
		console("initpkg called")
		next := vm.currentFrame.step.Next()
		vm.pushNewFrame(nil)
		vm.currentFrame.step = declare
		vm.currentFrame.returnTo = next
	})

	vm.pushNewFrame(nil)
	defer vm.popFrame()

	initPkg.SetNext(target)

	vm.currentFrame.step = initPkg
	for {
		if vm.currentFrame == nil {
			t.Log("no more frames, exiting")
			break
		}
		t.Log("step:", vm.currentFrame.step)
		err := vm.Next()
		if err != nil {
			if err != io.EOF {
				t.Fatalf("unexpected error: %v", err)
			}
			break
		}
	}
}
