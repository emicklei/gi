package pkg

// A runner can interpret a function step-by-step.
type Runner struct {
	all map[string]*Package
	vm  *VM
}

func NewRunner(mainPackage *Package) *Runner {
	r := &Runner{
		all: make(map[string]*Package),
		vm:  NewVM(mainPackage),
	}
	r.collectPackages(mainPackage)
	return r
}
func (r *Runner) collectPackages(from *Package) {
	if _, ok := r.all[from.PkgPath]; ok {
		return
	}
	r.all[from.PkgPath] = from
	for _, subpkg := range from.env.packageTable {
		r.collectPackages(subpkg)
	}
}
func (r *Runner) Step() error {
	return nil
}

func (r *Runner) Location() Step {
	return nil
}
