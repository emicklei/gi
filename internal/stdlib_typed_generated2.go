package internal

import (
	i1 "archive/tar"
	"reflect"
)

// stdtypes maps fully qualified type names to their exported types as reflect.Value not reflect.Type
var stdtypes2 = map[string]reflect.Value{}

// format: package-path.TypeName
func init() {
	stdtypes2["archive/tar.Format"] = makeReflect[i1.Format]()
}
