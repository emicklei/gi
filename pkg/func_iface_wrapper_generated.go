package pkg

import "reflect"

type fmtStringer struct {
	sdkInterfaceWrapper
}

// fmt.Stringer
func (s fmtStringer) String() string {
	decl, ok := s.typ.methodsMap()["String"]
	if !ok {
		s.vm.fatal("func declaration for 'String' not found")
	}
	s.vm.takeAllStartingAt(decl.graph)
	result := make([]reflect.Value, 1)
	result[0] = s.vm.popOperand()
	return result[0].Interface().(string)
}
