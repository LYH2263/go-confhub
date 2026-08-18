package confhub

import (
	"testing"
)

func TestBug01_PutCallerMutatesStoredPayload(t *testing.T) {
	h := New()
	t.Cleanup(func() { _ = h.Close() })
	if _, err := h.CreateNS("prod", "alice", 0, 0); err != nil {
		t.Fatal(err)
	}
	buf := []byte("alpha-config")
	if _, err := h.Put("prod", "app.msg", buf, PutOptions{Author: "alice"}); err != nil {
		t.Fatal(err)
	}
	buf[0] = 'Z'
	got, err := h.Get("prod", "app.msg", ClientContext{})
	if err != nil {
		t.Fatal(err)
	}
	if string(got.Payload) != "alpha-config" {
		t.Fatalf("Put must copy payload; after caller mutated buffer Get=%q", got.Payload)
	}
}
