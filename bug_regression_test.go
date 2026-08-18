package confhub

import "testing"

func TestBug10_DeleteNSChecksExistBeforeWipe(t *testing.T) {
	h := New()
	t.Cleanup(func() { _ = h.Close() })
	if _, err := h.CreateNS("prod", "alice", 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := h.CreateNS("other", "bob", 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Put("prod", "k", []byte("prod-val"), PutOptions{Author: "alice"}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Put("other", "k", []byte("other-val"), PutOptions{Author: "bob"}); err != nil {
		t.Fatal(err)
	}
	if err := h.DeleteNS("ghost"); err == nil {
		t.Fatal("missing namespace must error")
	}
	got, err := h.Get("prod", "k", ClientContext{})
	if err != nil {
		t.Fatalf("delete missing ns wiped prod: %v", err)
	}
	if string(got.Payload) != "prod-val" {
		t.Fatalf("prod payload=%q", got.Payload)
	}
	got, err = h.Get("other", "k", ClientContext{})
	if err != nil {
		t.Fatalf("delete missing ns wiped other: %v", err)
	}
	if string(got.Payload) != "other-val" {
		t.Fatalf("other payload=%q", got.Payload)
	}
	if err := h.DeleteNS("prod"); err != nil {
		t.Fatal(err)
	}
	got, err = h.Get("other", "k", ClientContext{})
	if err != nil {
		t.Fatalf("delete prod wiped other: %v", err)
	}
	if string(got.Payload) != "other-val" {
		t.Fatalf("other after prod delete: %q", got.Payload)
	}
}
