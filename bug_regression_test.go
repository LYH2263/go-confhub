package confhub

import (
	"testing"

	cherr "github.com/LYH2263/go-confhub/internal/errors"
	"github.com/LYH2263/go-confhub/internal/sign"
)

func TestBug04_FailedSignOrQuotaLeavesAppendedVersion(t *testing.T) {
	secret := []byte("super-secret-key")
	h := New(WithHMACSecret(secret), WithRequireSign(true))
	t.Cleanup(func() { _ = h.Close() })
	if _, err := h.CreateNS("prod", "alice", 0, 0); err != nil {
		t.Fatal(err)
	}
	_, err := h.Put("prod", "signed", []byte("ghost"), PutOptions{
		Author: "alice", Algo: sign.AlgoHMAC, Signature: []byte("deadbeef"),
	})
	if err == nil {
		t.Fatal("want signature failure")
	}
	if got, gerr := h.Get("prod", "signed", ClientContext{}); gerr == nil {
		t.Fatalf("failed sign must not publish; Get rev=%d payload=%q", got.Rev, got.Payload)
	}

	h2 := New()
	t.Cleanup(func() { _ = h2.Close() })
	if _, err := h2.CreateNS("tiny", "bob", 1, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := h2.Put("tiny", "a", []byte("keep"), PutOptions{Author: "bob"}); err != nil {
		t.Fatal(err)
	}
	_, err = h2.Put("tiny", "b", []byte("overflow"), PutOptions{Author: "bob"})
	if !cherr.Is(err, cherr.ErrQuotaKeys) {
		t.Fatalf("want key quota, got %v", err)
	}
	keys, err := h2.ListKeys("tiny")
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 || keys[0] != "a" {
		t.Fatalf("quota fail left extra key: %v", keys)
	}
	if _, gerr := h2.Get("tiny", "b", ClientContext{}); gerr == nil {
		t.Fatal("quota-fail Put must not leave key b visible")
	}
}
