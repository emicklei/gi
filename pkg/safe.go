package pkg

import (
	"fmt"
	"os"
	"reflect"
)

func init() {
	if os.Getenv("GI_IGNORE_EXIT") != "" {
		OnOsExit(func(code int) {
			fmt.Fprintf(os.Stderr, "[gi] os.Exit called with code %d\n", code)
		})
	}
	if os.Getenv("GI_IGNORE_PANIC") != "" {
		OnPanic(func(why any) {
			fmt.Fprintf(os.Stderr, "[gi] panic called with %v\n", why)
		})
	}
}

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
