package internal

import "testing"

func TestEnvPushPop(t *testing.T) {
	b := newStepBuilder(nil)
	top := b.env
	b.pushEnv()
	if b.env.getParent() != top {
		t.Errorf("after push, parent must be %v, got %v", top, b.env.getParent())
	}
	b.popEnv()
	if b.env != top {
		t.Errorf("after pop, env must be %v, got %v", top, b.env)
	}
}
