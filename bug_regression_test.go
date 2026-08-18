package confhub

import (
	"hash/fnv"
	"strconv"
	"testing"
)

func TestBug02_GrayMissReturnsStableNotHead(t *testing.T) {
	h := New()
	t.Cleanup(func() { _ = h.Close() })
	if _, err := h.CreateNS("prod", "alice", 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Put("prod", "flag", []byte("stable"), PutOptions{Author: "alice"}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Put("prod", "flag", []byte("canary"), PutOptions{
		Author: "alice",
		Gray:   &GrayRule{Percent: 10},
	}); err != nil {
		t.Fatal(err)
	}
	missID := clientWithBucket(t, 50)
	miss, err := h.Get("prod", "flag", ClientContext{ID: missID})
	if err != nil {
		t.Fatal(err)
	}
	if miss.Rev != 1 || miss.Hit || string(miss.Payload) != "stable" {
		t.Fatalf("gray miss must return Stable not Head: rev=%d hit=%v payload=%q", miss.Rev, miss.Hit, miss.Payload)
	}
	hitID := clientWithBucket(t, 0)
	hit, err := h.Get("prod", "flag", ClientContext{ID: hitID})
	if err != nil {
		t.Fatal(err)
	}
	if hit.Rev != 2 || !hit.Hit || string(hit.Payload) != "canary" {
		t.Fatalf("gray hit must return Head: rev=%d hit=%v payload=%q", hit.Rev, hit.Hit, hit.Payload)
	}
	if hit.Rev == miss.Rev {
		t.Fatalf("hit and miss must differ: both %d", hit.Rev)
	}
}

func bugBucket(id string) int {
	h := fnv.New64a()
	_, _ = h.Write([]byte(id))
	return int(h.Sum64() % 100)
}

func clientWithBucket(t *testing.T, want int) string {
	t.Helper()
	for i := 0; i < 200000; i++ {
		id := "c" + strconv.Itoa(i)
		if bugBucket(id) == want {
			return id
		}
	}
	t.Fatalf("no client with bucket %d", want)
	return ""
}
