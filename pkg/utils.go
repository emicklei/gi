package pkg

import (
	"fmt"
	"go/token"
	"io"
	"path/filepath"
	"reflect"
)

// reflectCondition converts a boolean to shared reflect.Value.
func reflectCondition(b bool) reflect.Value {
	if b {
		return reflectTrue
	}
	return reflectFalse
}

func isEllipsis(t Expr) bool {
	_, ok := t.(Ellipsis)
	return ok
}
func isStructValue(v reflect.Value) bool {
	_, ok := v.Interface().(StructValue)
	return ok
}

func isPointerExpr(e Expr) bool {
	_, ok := e.(StarExpr)
	return ok
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

func stringOf(v any) string {
	// Note: the order of tests is important!
	if v == undeclaredNil {
		return "<undeclared>"
	}
	if v == untypedNil {
		return "<untyped nil>"
	}
	if v == nil {
		return "nil"
	}
	if s, ok := v.(string); ok {
		return s
	}
	if rv, ok := v.(reflect.Value); ok {
		if rv.IsValid() && rv.CanInterface() {
			return stringOf(rv.Interface())
		} else {
			return fmt.Sprintf("%v", rv)
		}
	}
	if et, ok := v.(ExtendedValue); ok {
		if et.val.IsValid() {
			return fmt.Sprintf("%v", et.val.Interface())
		} else {
			return fmt.Sprintf("%v", et.val)
		}
	}
	if fs, ok := v.(fmt.Stringer); ok {
		return fs.String()
	}
	if psv, ok := v.(*StructValue); ok {
		return fmt.Sprintf("%p", psv)
	}
	if fm, ok := v.(fmt.Formatter); ok {
		return fmt.Sprintf("%v", fm)
	}
	if ts, ok := v.(ToStringer); ok {
		return ts.toString()
	}
	return fmt.Sprintf("%v", v)
}

func typeNameOf(_ any) string { return "any" }

func cursor(fs *token.FileSet, pos token.Pos) string {
	if pos == token.NoPos {
		return "<no position info>"
	}
	loc := fs.Position(pos)
	return fmt.Sprintf("%s:%d:%d", filepath.Base(loc.Filename), loc.Line, loc.Column)
}
