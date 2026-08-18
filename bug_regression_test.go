package confhub

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LYH2263/go-confhub/internal/meta"
)

func TestBug09_RollbackSucceedsWithoutAuditOrPersist(t *testing.T) {
	h := New()
	t.Cleanup(func() { _ = h.Close() })
	if _, err := h.CreateNS("prod", "alice", 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Put("prod", "k", []byte("one"), PutOptions{Author: "alice"}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Put("prod", "k", []byte("two"), PutOptions{Author: "alice"}); err != nil {
		t.Fatal(err)
	}

	bad := t.TempDir()
	h.persistPath = bad
	if err := h.Rollback("prod", "k", 1); err == nil {
		t.Fatal("Rollback must surface persist error when snapshot path is a directory")
	}

	path := filepath.Join(t.TempDir(), "hub.json")
	h2 := New(WithPersistPath(path))
	t.Cleanup(func() { _ = h2.Close() })
	if _, err := h2.CreateNS("prod", "alice", 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := h2.Put("prod", "k", []byte("one"), PutOptions{Author: "alice"}); err != nil {
		t.Fatal(err)
	}
	if _, err := h2.Put("prod", "k", []byte("two"), PutOptions{Author: "alice"}); err != nil {
		t.Fatal(err)
	}
	if err := h2.Rollback("prod", "k", 1); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, rec := range h2.AuditQuery("prod", "k", 50) {
		if rec.Kind == meta.KindRollback && rec.OK && rec.Rev == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("successful Rollback must append an audit record")
	}
	st, err := os.Stat(path)
	if err != nil || st.Size() == 0 {
		t.Fatalf("successful Rollback must write persist snapshot: %v", err)
	}
}
