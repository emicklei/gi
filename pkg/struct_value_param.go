package pkg

type StructValueParam struct {
	vm    *VM
	value StructValue
}

func newStructValueParam(vm *VM, value StructValue) StructValueParam {
	return StructValueParam{vm: vm, value: value}
}
