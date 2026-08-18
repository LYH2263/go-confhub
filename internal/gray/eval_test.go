package gray

import (
	"testing"

	"github.com/LYH2263/go-confhub/internal/meta"
)

func TestEvalAllowListAndTags(t *testing.T) {
	rule := &Rule{
		Percent:   0,
		AllowIDs:  []string{"vip"},
		MatchTags: map[string]string{"env": "canary"},
	}
	if !Eval(rule, meta.ClientContext{ID: "vip"}) {
		t.Fatal("allowlist should hit even without tags")
	}
	if Eval(rule, meta.ClientContext{ID: "other", Tags: map[string]string{"env": "canary"}}) {
		t.Fatal("percent 0 without allow should miss even if tags match")
	}
	rule.Percent = 100
	if !Eval(rule, meta.ClientContext{ID: "other", Tags: map[string]string{"env": "canary"}}) {
		t.Fatal("percent 100 with matching tags should hit")
	}
	if Eval(rule, meta.ClientContext{ID: "other", Tags: map[string]string{"env": "prod"}}) {
		t.Fatal("tag mismatch should miss")
	}
}

func TestBucketStable(t *testing.T) {
	a := Bucket("client-a")
	b := Bucket("client-a")
	if a != b || a < 0 || a > 99 {
		t.Fatalf("bucket %d %d", a, b)
	}
}

func TestSelectNoStable(t *testing.T) {
	e := &meta.Entry{
		NS: "n", Key: "k", Head: 1, Stable: 0,
		Versions: []meta.VersionMeta{{Rev: 1, Payload: []byte("x"), Gray: &meta.GrayRule{Percent: 1, AllowIDs: []string{"only"}}}},
	}
	_, _, ok := Select(e, meta.ClientContext{ID: "nope"})
	if ok {
		t.Fatal("miss without stable must not be ok")
	}
	rev, hit, ok := Select(e, meta.ClientContext{ID: "only"})
	if !ok || !hit || rev != 1 {
		t.Fatalf("allow hit: %d %v %v", rev, hit, ok)
	}
}
