package pkg

import (
	"reflect"
	"sync"
)

var stdfuncsMutex sync.Mutex

func replaceStdFunc(pkg, name string, f reflect.Value) {
	stdfuncsMutex.Lock()
	defer stdfuncsMutex.Unlock()
	stdfuncs[pkg][name] = f
}

// OnPanic sets the function to be called when panic is invoked in the interpreted code.
// The Go SDK panic is not called.
func OnPanic(f func(any)) {
	// TODO make this thread-safe
	builtins["panic"] = reflect.ValueOf(f)
}

// OnOsExit sets the function to be called when os.Exit is invoked in the interpreted code.
// The Go SDK os.Exit is not called.
func OnOsExit(f func(int)) {
	replaceStdFunc("os", "Exit", reflect.ValueOf(f))
}
