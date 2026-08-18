package quota

import (
	"testing"

	cherr "github.com/LYH2263/go-confhub/internal/errors"
	"github.com/LYH2263/go-confhub/internal/meta"
)

func TestLimiterCheckThenCommit(t *testing.T) {
	l := NewLimiter()
	ns := meta.Namespace{ID: "n", MaxKeys: 2, MaxBytes: 10}
	if err := l.Check(CheckInput{NS: ns, NewKey: true, AddBytes: 4}); err != nil {
		t.Fatal(err)
	}
	l.Commit("n", 1, 4)
	if err := l.Check(CheckInput{NS: ns, NewKey: true, AddBytes: 4}); err != nil {
		t.Fatal(err)
	}
	l.Commit("n", 1, 4)
	err := l.Check(CheckInput{NS: ns, NewKey: true, AddBytes: 1})
	if !cherr.Is(err, cherr.ErrQuotaKeys) {
		t.Fatalf("got %v", err)
	}
	err = l.Check(CheckInput{NS: ns, NewKey: false, AddBytes: 3})
	if !cherr.Is(err, cherr.ErrQuotaBytes) {
		t.Fatalf("got %v", err)
	}
	u := l.Usage("n")
	if u.Keys != 2 || u.Bytes != 8 {
		t.Fatalf("usage mutated on failed check: %+v", u)
	}
}
