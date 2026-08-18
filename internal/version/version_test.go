package version

import "testing"

func TestDiffAndParse(t *testing.T) {
	d := Diff([]byte("a\nb\n"), []byte("a\nc\n"))
	if d.Equal || d.Binary {
		t.Fatalf("%+v", d)
	}
	u := Unified(d)
	if u == "" {
		t.Fatal("expected unified diff")
	}
	n, err := Parse("r12")
	if err != nil || n != 12 {
		t.Fatalf("%v %v", n, err)
	}
	if err := CheckStrictInc([]int64{1, 2, 4}); err == nil {
		t.Fatal("gap")
	}
}
