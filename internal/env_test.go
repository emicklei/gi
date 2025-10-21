package internal

import (
	"reflect"
	"testing"
)

func TestPackageEnv(t *testing.T) {
	global := newEnvironment(nil)
	global.set("a", reflect.ValueOf(1))
	pkg := newPkgEnvironment(global)
	fn := pkg.newChild()
	one := fn.valueLookUp("a")
	if one.Interface() != 1 {
		t.Fail()
	}
	sub := pkg.newChild()
	if sub.getParent() != pkg {
		t.Errorf("sub's parent must be pkg, got %v, want %v", sub.getParent(), pkg)
	}
	if pkg.getParent() != global {
		t.Errorf("pkg's parent must be global, got %v, want %v", pkg.getParent(), global)
	}
}
