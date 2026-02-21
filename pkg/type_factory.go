package pkg

import "reflect"

// never returns a CanMake
func makeType(vm *VM, e Evaluable) reflect.Type {
	if id, ok := e.(Ident); ok {
		typ, ok := builtins[id.name]
		if ok {
			return typ.Interface().(builtinType).typ
		}
		return structValueType
	}
	if star, ok := e.(StarExpr); ok {
		nonStarType := makeType(vm, star.x)
		return reflect.PointerTo(nonStarType)
	}
	if sel, ok := e.(SelectorExpr); ok {
		typ := vm.currentEnv().valueLookUp(sel.x.(Ident).name)
		val := typ.Interface()
		if canSelect, ok := val.(CanSelect); ok {
			selVal := canSelect.selectFieldOrMethod(sel.selector.name)
			return reflect.TypeOf(selVal.Interface())
		}
		pkgType := stdtypes[sel.x.(Ident).name][sel.selector.name]
		return reflect.TypeOf(pkgType.Interface())
	}
	if ar, ok := e.(ArrayType); ok {
		elemType := makeType(vm, ar.elt)
		if ar.len == nil {
			return reflect.SliceOf(elemType)
		} else {
			lenVal := vm.returnsEval(ar.len)
			size := int(lenVal.Int())
			return reflect.ArrayOf(size, elemType)
		}
	}
	if _, ok := e.(FuncType); ok {
		// any function type will do; we just need its reflect.Type
		fn := func() {}
		return reflect.TypeOf(fn)
	}
	if e, ok := e.(Ellipsis); ok {
		return makeType(vm, e.elt)
	}
	vm.fatalf("unhandled makeType for %v (%T)", e, e)
	return nil
}

func typeMaker(vm *VM, e Expr) CanMake {
	if id, ok := e.(Ident); ok {
		typ, ok := builtins[id.name]
		if ok {
			gt := SDKType{typ: typ.Interface().(builtinType).typ}
			return gt
		}
		typ = vm.currentEnv().valueLookUp(id.name)
		// interpreted
		if cm, ok := typ.Interface().(CanMake); ok {
			return cm
		}
		vm.fatalf("unhandled proxyType for %v (%T)", e, e)
	}

	if sel, ok := e.(SelectorExpr); ok {
		typ := vm.currentEnv().valueLookUp(sel.x.(Ident).name)
		val := typ.Interface()
		if canSelect, ok := val.(CanSelect); ok {
			selVal := canSelect.selectFieldOrMethod(sel.selector.name)
			return SDKType{typ: reflect.TypeOf(selVal.Interface())}
		}
		pkgType := stdtypes[sel.x.(Ident).name][sel.selector.name]
		return SDKType{typ: reflect.TypeOf(pkgType.Interface())}
	}

	if star, ok := e.(StarExpr); ok {
		nonStarType := typeMaker(vm, star.x)
		return nonStarType // .pointerType(). // TODO
	}

	if ar, ok := e.(ArrayType); ok {
		elemType := makeType(vm, ar.elt)
		if ar.len == nil {
			return SDKType{typ: reflect.SliceOf(elemType)}
		} else {
			lenVal := vm.returnsEval(ar.len)
			size := int(lenVal.Int())
			return SDKType{typ: reflect.ArrayOf(size, elemType)}
		}
	}

	if _, ok := e.(FuncType); ok {
		// any function type will do; we just need its reflect.Type
		// TODO
		fn := func() {}
		return SDKType{typ: reflect.TypeOf(fn)}
	}

	if e, ok := e.(Ellipsis); ok {
		return typeMaker(vm, e.elt)
	}

	vm.fatalf("unhandled typeMaker for %v (%T)", e, e)
	return nil
}
