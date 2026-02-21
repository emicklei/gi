package pkg

import (
	"reflect"
	"testing"
)

func TestPackageEnv(t *testing.T) {
	global := newEnvironment(nil)
	global.valueSet("a", reflect.ValueOf(1))
	pkg := newPkgEnvironment(global)
	fn := pkg.newChild()
	one := fn.valueLookUp("a")
	if one.Interface() != 1 {
		t.Fail()
	}
	sub := pkg.newChild()
	if sub.parent() != pkg {
		t.Errorf("sub's parent must be pkg, got %v, want %v", sub.parent(), pkg)
	}
	if pkg.parent() != global {
		t.Errorf("pkg's parent must be global, got %v, want %v", pkg.parent(), global)
	}
}
