package confhub

import (
	"testing"

	cherr "github.com/LYH2263/go-confhub/internal/errors"
)

func TestBug07_GetAfterCloseReturnsTypedError(t *testing.T) {
	h := New()
	if _, err := h.CreateNS("prod", "alice", 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Put("prod", "app.msg", []byte("live"), PutOptions{Author: "alice"}); err != nil {
		t.Fatal(err)
	}
	if err := h.Close(); err != nil {
		t.Fatal(err)
	}
	_, err := h.Get("prod", "app.msg", ClientContext{})
	if err == nil {
		t.Fatal("Get after Close must fail")
	}
	if !cherr.Is(err, cherr.ErrClosed) {
		t.Fatalf("Get after Close want ErrClosed, got %v", err)
	}
}
