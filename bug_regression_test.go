package confhub_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LYH2263/go-confhub"
	"github.com/LYH2263/go-confhub/internal/api"
)

func TestBug08_HTTPWatchIgnoresRequestContext(t *testing.T) {
	h := confhub.New()
	t.Cleanup(func() { _ = h.Close() })
	if _, err := h.CreateNS("prod", "alice", 0, 0); err != nil {
		t.Fatal(err)
	}
	s := api.New(h, api.Options{})
	ts := httptest.NewServer(s)
	t.Cleanup(ts.Close)

	ctx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/api/ns/prod/keys/k/watch?timeout=15s&since=0", nil)
	if err != nil {
		t.Fatal(err)
	}
	errc := make(chan error, 1)
	go func() {
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			errc <- err
			return
		}
		defer res.Body.Close()
		_, _ = io.ReadAll(res.Body)
		errc <- nil
	}()

	deadline := time.Now().Add(2 * time.Second)
	for h.WatchHub().Count() < 1 {
		if time.Now().After(deadline) {
			t.Fatal("watch handler did not subscribe")
		}
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	deadline = time.Now().Add(800 * time.Millisecond)
	for time.Now().Before(deadline) {
		if h.WatchHub().Count() == 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if _, err := h.Put("prod", "k", []byte("late"), confhub.PutOptions{Author: "alice"}); err != nil {
		t.Fatal(err)
	}
	if n := h.WatchHub().Count(); n != 0 {
		t.Fatalf("canceled HTTP watch still subscribed after Publish: count=%d", n)
	}
	select {
	case <-errc:
	case <-time.After(2 * time.Second):
		t.Fatal("watch handler did not return after request cancel")
	}
}
