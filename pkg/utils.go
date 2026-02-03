package pkg

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

// pkgIdent = fully (package) qualified (identifying) name
type pkgIdent struct {
	path string
	name string
}

// Deprecated
func expected(value any, expectation string) reflect.Value {
	panic(fmt.Sprintf("expected %s : %v (%T)", expectation, value, value))
}

func mustString(v reflect.Value) string {
	s, ok := v.Interface().(string)
	if ok {
		return s
	}
	i, ok := v.Interface().(Ident)
	if ok {
		return i.name
	}
	panic(fmt.Sprintf("expected string or undeclaredbut got %T", v.Interface()))
}

func mustIdentName(e Expr) string {
	if id, ok := e.(Ident); ok {
		return id.name
	}
	if id, ok := e.(*Ident); ok {
		return id.name
	}
	panic(fmt.Sprintf("expected Ident but got %T", e))
}

// Push return values onto the operand stack in reverse order,
// so the first return value ends up on top of the stack.
func pushCallResults(vm *VM, vals []reflect.Value) {
	for i := len(vals) - 1; i >= 0; i-- {
		vm.pushOperand(vals[i])
	}
}

// makeReflect is a helper function used in generated code to create reflect.Value of a type T.
func makeReflect[T any]() reflect.Value {
	var t T
	return reflect.ValueOf(t)
}

// format writes a string representation of the value 'val' to the writer 'w',
// according to the formatting verb.
func format(w io.Writer, verb rune, val any) {
	// First, handle the nil case, which is important for any interface value.
	if val == nil {
		io.WriteString(w, "<nil>")
		return
	}

	// The main type switch to determine the concrete type of 'val'.
	switch v := val.(type) {
	// --- Interface Checks ---
	// Check for fmt.Formatter first, as it's the most specific.
	case fmt.Formatter:
		// The type knows how to format itself. Delegate the work to it.
		// We pass a 'state' that wraps our writer and the verb.
		v.Format(w.(fmt.State), verb)

	// If not a Formatter, check if it's a Stringer.
	case fmt.Stringer:
		// The type can represent itself as a simple string.
		// This is good for verbs 's' and 'v'.
		switch verb {
		case 's', 'v':
			io.WriteString(w, v.String())
		default:
			// For other verbs, we indicate an error.
			fmt.Fprintf(w, "%%!%c(%T=%s)", verb, v, v.String())
		}

	// --- Concrete Type Handlers ---
	case string:
		switch verb {
		case 's', 'v':
			io.WriteString(w, v)
		case 'q':
			fmt.Fprintf(w, "%q", v)
		default:
			fmt.Fprintf(w, "%%!%c(string=%s)", verb, v)
		}

	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		// Handle all integer types.
		switch verb {
		case 'd', 'v': // decimal
			fmt.Fprintf(w, "%d", v)
		case 'b': // binary
			fmt.Fprintf(w, "%b", v)
		case 'o': // octal
			fmt.Fprintf(w, "%o", v)
		case 'x', 'X': // hex
			// Pass the verb through to respect case.
			formatStr := "%" + string(verb)
			fmt.Fprintf(w, formatStr, v)
		default:
			fmt.Fprintf(w, "%%!%c(%T=%d)", verb, v, v)
		}

	case bool:
		switch verb {
		case 't', 'v':
			fmt.Fprintf(w, "%t", v)
		default:
			fmt.Fprintf(w, "%%!%c(bool=%t)", verb, v)
		}

	case float32, float64:
		// Handle all float types.
		switch verb {
		case 'f', 'g', 'e', 'E', 'v':
			// Pass the verb through.
			formatStr := "%" + string(verb)
			fmt.Fprintf(w, formatStr, v)
		default:
			fmt.Fprintf(w, "%%!%c(%T=%g)", verb, v, v)
		}

	// --- The Fallback Case ---
	default:
		// For any other type not handled above (slices, maps, structs, etc.),
		// fall back to the default fmt behavior.
		formatStr := "%" + string(verb)
		fmt.Fprintf(w, formatStr, v)
	}
}

// prints types and values
func console(v any) {
	if rt, ok := v.(reflect.Type); ok {
		fmt.Printf("console: type: %s,%v\n", rt.Name(), rt)
		return
	}
	if ts, ok := v.(ToStringer); ok {
		fmt.Printf("console: %s\n", ts.toString())
		return
	}
	// fallback
	fmt.Printf("console: %#v (%T)\n", v, v)
}

func internalVarName(meaning string, seq int) string {
	return fmt.Sprintf("_%s_%d", meaning, seq)
}

func fieldTypeExpr(fields *FieldList, index int) Expr {
	count := 0
	for _, field := range fields.List {
		for range field.names {
			if count == index {
				return field.typ
			}
			count++
		}
	}
	return nil
}

func isPointerToStructValue(v reflect.Value) bool {
	if v.Kind() != reflect.Pointer {
		return false
	}
	if v.Elem().Kind() != reflect.Struct {
		return false
	}
	if v.Elem().Type().Name() != "StructValue" {
		return false
	}
	// not exact package match to allow source code forks
	if !strings.HasSuffix(v.Elem().Type().PkgPath(), "/gi/pkg") {
		return false
	}
	return true
}
