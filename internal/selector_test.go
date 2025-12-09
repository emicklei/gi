package internal

import (
	"reflect"
	"sync"
	"testing"
)

func TestPointerMethodWaitGroup(t *testing.T) {
	var wg sync.WaitGroup

	twg := reflect.TypeOf(wg)
	meth, ok := twg.MethodByName("Go")
	t.Log(meth, ok)

	ptwg := reflect.PointerTo(twg)

	meth, ok = ptwg.MethodByName("Go")
	t.Log(meth, ok)

}
