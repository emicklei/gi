package pkg

import (
	"testing"

	"github.com/google/go-dap"
)

func TestDAPAccess(t *testing.T) {
	src := `package main
func main(){
	msg := "gi"
	print(msg)
}`
	pkg := buildPackage(t, src)
	vm := NewVM(pkg)
	xs := NewDAPAccess(vm)
	xs.Launch("main", nil)
	for {
		err := xs.Next()
		if err != nil {
			t.Log(err)
			break
		}
		for _, thread := range xs.Threads() {
			for _, frame := range xs.StackFrames(dap.StackTraceArguments{ThreadId: thread.Id}) {
				for _, scope := range xs.Scopes(dap.ScopesArguments{FrameId: frame.Id}) {
					for _, v := range xs.Variables(dap.VariablesArguments{VariablesReference: scope.VariablesReference}) {
						t.Logf("thread %d frame %d: %s = %v", thread.Id, frame.Id, v.Name, v.Value)
					}
				}
			}
		}
	}
}
