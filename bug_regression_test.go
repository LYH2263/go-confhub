package confhub

import (
	"context"
	"testing"
	"time"
)

func TestBug03_WatchWaitIgnoresContextCancel(t *testing.T) {
	h := New()
	t.Cleanup(func() { _ = h.Close() })
	if _, err := h.CreateNS("prod", "alice", 0, 0); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	ch, unsub := h.WatchHub().SubscribeContext(ctx, "prod", "k")
	defer unsub()
	cancel()
	time.Sleep(30 * time.Millisecond)
	if _, err := h.Put("prod", "k", []byte("after-cancel"), PutOptions{Author: "alice"}); err != nil {
		t.Fatal(err)
	}
	select {
	case ev, ok := <-ch:
		if ok {
			t.Fatalf("SubscribeContext delivered after ctx cancel: %+v", ev)
		}
	case <-time.After(80 * time.Millisecond):
		t.Fatal("canceled watch sub must close; plant still delivers")
	}

	ctx2, cancel2 := context.WithCancel(context.Background())
	errc := make(chan error, 1)
	go func() {
		ev, ok := h.WatchHub().WaitContext(ctx2, "prod", "k", 1, 3*time.Second)
		if ok {
			errc <- errWaitDeliver{evRev: ev.Rev}
			return
		}
		errc <- nil
	}()
	deadline := time.Now().Add(time.Second)
	for h.WatchHub().Count() < 1 {
		if time.Now().After(deadline) {
			t.Fatal("WaitContext did not subscribe")
		}
		time.Sleep(5 * time.Millisecond)
	}
	cancel2()
	if _, err := h.Put("prod", "k", []byte("wait-late"), PutOptions{Author: "alice"}); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-errc:
		if err != nil {
			t.Fatalf("WaitContext delivered after cancel: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("WaitContext ignored ctx cancel and kept blocking")
	}
}

type errWaitDeliver struct{ evRev int64 }

func (e errWaitDeliver) Error() string { return "wait delivered after cancel" }
