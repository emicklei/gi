package internal

import "testing"

func TestLoadGomod(t *testing.T) {
	modFile, err := LoadGoMod("../go.mod")
	if err != nil {
		t.Fatalf("failed to load go.mod: %v", err)
	}
	for _, each := range modFile.Require {
		t.Logf("require: %s %s", each.Mod.Path, each.Mod.Version)
		path, err := GetModulePath(each.Mod.Path, each.Mod.Version)
		t.Logf("module path: %s %v", path, err)
	}
}
