package confhub

import (
	"testing"
	"time"
)

func TestBug05_WatchWaitStrictAfterSinceAndMatchKey(t *testing.T) {
	h := New()
	t.Cleanup(func() { _ = h.Close() })
	if _, err := h.CreateNS("prod", "alice", 0, 0); err != nil {
		t.Fatal(err)
	}
	ch, cancel := h.Watch("prod", "only-this")
	t.Cleanup(cancel)
	if _, err := h.Put("prod", "other-key", []byte("nope"), PutOptions{Author: "alice"}); err != nil {
		t.Fatal(err)
	}
	select {
	case ev := <-ch:
		t.Fatalf("Watch must Match key, got event for %s/%s rev=%d", ev.NS, ev.Key, ev.Rev)
	case <-time.After(80 * time.Millisecond):
	}
	if _, err := h.Put("prod", "only-this", []byte("one"), PutOptions{Author: "alice"}); err != nil {
		t.Fatal(err)
	}
	select {
	case ev := <-ch:
		if ev.Key != "only-this" || ev.Rev != 1 {
			t.Fatalf("want only-this rev=1, got %+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("watch not woken for matching key")
	}
	ev, ok := h.WatchHub().Wait("prod", "only-this", 1, 80*time.Millisecond)
	if ok {
		t.Fatalf("Wait must only return Rev>since, got %+v", ev)
	}
}
