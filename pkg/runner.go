package pkg

import "slices"

// A runner can interpret a function step-by-step.
type Runner struct {
	all []*Package // ordered and unique
	vm  *VM
}

func NewRunner(mainPackage *Package) *Runner {
	r := &Runner{
		all: []*Package{},
		vm:  NewVM(mainPackage),
	}
	r.collectPackages(mainPackage)
	return r
}
func (r *Runner) collectPackages(from *Package) {
	if slices.Contains(r.all, from) {
		return
	}
	r.all = append(r.all, from)
	for _, subpkg := range from.env.packageTable {
		r.collectPackages(subpkg)
	}
}
func (r *Runner) Step() error {
	// for now skip initialization
	call := CallExpr{
		fun: Ident{name: "main"},
	}
	fun := r.vm.localEnv().valueLookUp("main")
	r.vm.pushOperand(fun)
	r.vm.eval(call)
	return nil
}

func (r *Runner) Location() Step {
	return nil
}
