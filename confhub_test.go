package confhub_test

import (
	"testing"
	"time"

	"github.com/LYH2263/go-confhub"
	cherr "github.com/LYH2263/go-confhub/internal/errors"
	"github.com/LYH2263/go-confhub/internal/gray"
	"github.com/LYH2263/go-confhub/internal/meta"
	"github.com/LYH2263/go-confhub/internal/sign"
	"github.com/LYH2263/go-confhub/internal/version"
)

func TestVersionStrictIncrement(t *testing.T) {
	h := confhub.New()
	t.Cleanup(func() { _ = h.Close() })
	if _, err := h.CreateNS("prod", "alice", 100, 1<<20); err != nil {
		t.Fatal(err)
	}
	var revs []int64
	for i := 0; i < 5; i++ {
		rev, err := h.Put("prod", "app.timeout", []byte("v"+string(rune('0'+i))), confhub.PutOptions{Author: "alice"})
		if err != nil {
			t.Fatal(err)
		}
		revs = append(revs, rev)
	}
	if err := version.CheckStrictInc(revs); err != nil {
		t.Fatal(err)
	}
	got, err := h.Get("prod", "app.timeout", confhub.ClientContext{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Rev != 5 || got.Head != 5 {
		t.Fatalf("latest jumped: %+v", got)
	}
}

func TestGrayHitMissDifferentPointers(t *testing.T) {
	h := confhub.New()
	t.Cleanup(func() { _ = h.Close() })
	if _, err := h.CreateNS("prod", "alice", 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Put("prod", "flag", []byte("stable"), confhub.PutOptions{Author: "alice"}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Put("prod", "flag", []byte("canary"), confhub.PutOptions{
		Author: "alice",
		Gray:   &meta.GrayRule{Percent: 10},
	}); err != nil {
		t.Fatal(err)
	}
	var hitID, missID string
	for i := 0; i < 500; i++ {
		id := "c" + itoa(i)
		if gray.InPercent(id, 10) {
			hitID = id
		} else {
			missID = id
		}
		if hitID != "" && missID != "" {
			break
		}
	}
	if hitID == "" || missID == "" {
		t.Fatal("could not find hit/miss clients")
	}
	hit, err := h.Get("prod", "flag", confhub.ClientContext{ID: hitID})
	if err != nil {
		t.Fatal(err)
	}
	miss, err := h.Get("prod", "flag", confhub.ClientContext{ID: missID})
	if err != nil {
		t.Fatal(err)
	}
	if hit.Rev == miss.Rev {
		t.Fatalf("gray hit/miss must differ: hit=%d miss=%d", hit.Rev, miss.Rev)
	}
	if hit.Rev != 2 || miss.Rev != 1 {
		t.Fatalf("unexpected revs hit=%d miss=%d", hit.Rev, miss.Rev)
	}
	if !hit.Hit || miss.Hit {
		t.Fatalf("hit flags: hit=%v miss=%v", hit.Hit, miss.Hit)
	}
}

func TestWatchWakeAndCancel(t *testing.T) {
	h := confhub.New()
	t.Cleanup(func() { _ = h.Close() })
	if _, err := h.CreateNS("prod", "alice", 0, 0); err != nil {
		t.Fatal(err)
	}
	ch, cancel := h.Watch("prod", "k")
	putErr := make(chan error, 1)
	go func() {
		time.Sleep(20 * time.Millisecond)
		_, err := h.Put("prod", "k", []byte("one"), confhub.PutOptions{Author: "alice"})
		putErr <- err
	}()
	select {
	case ev := <-ch:
		if ev.Kind != meta.KindPublish || ev.Rev != 1 {
			t.Fatalf("event: %+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("watch not woken")
	}
	if err := <-putErr; err != nil {
		t.Fatal(err)
	}
	cancel()
	if _, ok := <-ch; ok {
		t.Fatal("expected channel closed after cancel")
	}
	if _, err := h.Put("prod", "k", []byte("two"), confhub.PutOptions{Author: "alice"}); err != nil {
		t.Fatal(err)
	}
	select {
	case ev, ok := <-ch:
		if ok {
			t.Fatalf("delivered after cancel: %+v", ev)
		}
	case <-time.After(50 * time.Millisecond):
	}
}

func TestSignFailNotPublished(t *testing.T) {
	secret := []byte("super-secret-key")
	h := confhub.New(confhub.WithHMACSecret(secret), confhub.WithRequireSign(true))
	t.Cleanup(func() { _ = h.Close() })
	if _, err := h.CreateNS("prod", "alice", 0, 0); err != nil {
		t.Fatal(err)
	}
	payload := []byte("signed-body")
	good, err := h.SignPayload("prod", "k", "", sign.AlgoHMAC, 0, payload)
	if err != nil {
		t.Fatal(err)
	}
	_, err = h.Put("prod", "k", payload, confhub.PutOptions{
		Author: "alice", Algo: sign.AlgoHMAC, Signature: []byte("deadbeef"),
	})
	if !cherr.Is(err, cherr.ErrSignFailed) && err == nil {
		t.Fatalf("want sign fail, got %v", err)
	}
	if _, err := h.Get("prod", "k", confhub.ClientContext{}); err == nil {
		t.Fatal("unsigned/failed must not be published")
	}
	rev, err := h.Put("prod", "k", payload, confhub.PutOptions{
		Author: "alice", Algo: sign.AlgoHMAC, Signature: good,
	})
	if err != nil {
		t.Fatal(err)
	}
	if rev != 1 {
		t.Fatalf("rev=%d", rev)
	}
}

func TestQuotaNoPartialVersion(t *testing.T) {
	h := confhub.New()
	t.Cleanup(func() { _ = h.Close() })
	if _, err := h.CreateNS("tiny", "bob", 1, 16); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Put("tiny", "a", []byte("1234"), confhub.PutOptions{Author: "bob"}); err != nil {
		t.Fatal(err)
	}
	_, err := h.Put("tiny", "b", []byte("xxxx"), confhub.PutOptions{Author: "bob"})
	if !cherr.Is(err, cherr.ErrQuotaKeys) {
		t.Fatalf("want key quota, got %v", err)
	}
	keys, err := h.ListKeys("tiny")
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 || keys[0] != "a" {
		t.Fatalf("partial key left: %v", keys)
	}
	_, err = h.Put("tiny", "a", []byte("0123456789abcdef0123"), confhub.PutOptions{Author: "bob"})
	if !cherr.Is(err, cherr.ErrQuotaBytes) {
		t.Fatalf("want byte quota, got %v", err)
	}
	e, err := h.GetEntry("tiny", "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(e.Versions) != 1 {
		t.Fatalf("partial version left: %d", len(e.Versions))
	}
}

func TestRollbackKeepsHistory(t *testing.T) {
	h := confhub.New()
	t.Cleanup(func() { _ = h.Close() })
	if _, err := h.CreateNS("prod", "alice", 0, 0); err != nil {
		t.Fatal(err)
	}
	_, _ = h.Put("prod", "k", []byte("one"), confhub.PutOptions{})
	_, _ = h.Put("prod", "k", []byte("two"), confhub.PutOptions{})
	if err := h.Rollback("prod", "k", 1); err != nil {
		t.Fatal(err)
	}
	got, err := h.Get("prod", "k", confhub.ClientContext{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Rev != 1 || string(got.Payload) != "one" {
		t.Fatalf("rollback: %+v %q", got, got.Payload)
	}
	hist, err := h.History("prod", "k")
	if err != nil || len(hist) != 2 {
		t.Fatalf("history should keep both revs: %v %v", hist, err)
	}
}

func TestSnapshotRoundTrip(t *testing.T) {
	h := confhub.New()
	t.Cleanup(func() { _ = h.Close() })
	if _, err := h.CreateNS("prod", "alice", 8, 1024); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Put("prod", "k", []byte("hello"), confhub.PutOptions{Author: "alice"}); err != nil {
		t.Fatal(err)
	}
	raw, err := h.ExportBytes()
	if err != nil {
		t.Fatal(err)
	}
	h2 := confhub.New()
	t.Cleanup(func() { _ = h2.Close() })
	if err := h2.ImportBytes(raw); err != nil {
		t.Fatal(err)
	}
	got, err := h2.Get("prod", "k", confhub.ClientContext{})
	if err != nil {
		t.Fatal(err)
	}
	if string(got.Payload) != "hello" {
		t.Fatalf("got %q", got.Payload)
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [16]byte
	n := len(b)
	for i > 0 {
		n--
		b[n] = byte('0' + i%10)
		i /= 10
	}
	return string(b[n:])
}
