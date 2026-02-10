package pkg

import (
	"reflect"
)

// OnPanic sets the function to be called when panic is invoked in the interpreted code.
// The Go SDK panic is not called.
func OnPanic(f func(any)) {
	// TODO make this thread-safe
	builtinsMap["panic"] = reflect.ValueOf(f)
}

// OnOsExit sets the function to be called when os.Exit is invoked in the interpreted code.
// The Go SDK os.Exit is not called.
func OnOsExit(f func(int)) {
	// TODO make this thread-safe
	stdfuncs["os"]["Exit"] = reflect.ValueOf(f)
}
