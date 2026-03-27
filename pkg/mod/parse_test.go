package mod

import (
	"os"
	"testing"
)

func TestParseWriteDot(t *testing.T) {
	pc, err := ParsePackageContent("github.com/emicklei/dot")
	if err != nil {
		t.Error(err)
	}
	if len(pc.Values) == 0 {
		t.Fail()
	}
	if len(pc.Types) == 0 {
		t.Fail()
	}
	if err := pc.write(os.Stdout); err != nil {
		t.Fatal(err)
	}
}
