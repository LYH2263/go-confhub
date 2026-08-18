package watch

import (
	"testing"
	"time"
)

func TestHubCancelStopsDelivery(t *testing.T) {
	h := NewHub()
	ch, cancel := h.Subscribe("ns", "k")
	h.Publish(NewEvent("ns", "k", KindPublish, 1))
	select {
	case ev := <-ch:
		if ev.Rev != 1 {
			t.Fatalf("%+v", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("no event")
	}
	cancel()
	h.Publish(NewEvent("ns", "k", KindPublish, 2))
	select {
	case ev, ok := <-ch:
		if ok {
			t.Fatalf("got %v", ev)
		}
	case <-time.After(80 * time.Millisecond):
		t.Fatal("channel should be closed")
	}
}

func TestWaitLongPoll(t *testing.T) {
	h := NewHub()
	go func() {
		time.Sleep(30 * time.Millisecond)
		h.Publish(NewEvent("ns", "k", KindPublish, 3))
	}()
	ev, ok := h.Wait("ns", "k", 0, time.Second)
	if !ok || ev.Rev != 3 {
		t.Fatalf("wait: %+v %v", ev, ok)
	}
}
