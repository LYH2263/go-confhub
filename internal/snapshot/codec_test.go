package snapshot

import (
	"testing"
	"time"

	"github.com/LYH2263/go-confhub/internal/meta"
)

func TestBinaryCodecRoundTrip(t *testing.T) {
	b := Build(Source{
		Now: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		Namespaces: []meta.Namespace{{
			ID: "prod", Owner: "a", CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			MaxKeys: 10, MaxBytes: 100,
		}},
		Entries: []*meta.Entry{{
			NS: "prod", Key: "k", Head: 1, Stable: 1,
			Versions: []meta.VersionMeta{{
				Rev: 1, Payload: []byte("hi"), Author: "a",
				Ts:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				Gray: &meta.GrayRule{Percent: 5, AllowIDs: []string{"x"}},
			}},
		}},
	})
	raw, err := Encode(b)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Namespaces) != 1 || len(got.Entries) != 1 {
		t.Fatalf("%+v", got.Header)
	}
	if string(got.Entries[0].Versions[0].Payload) != "hi" {
		t.Fatal("payload")
	}
	if got.Entries[0].Versions[0].Gray.Percent != 5 {
		t.Fatal("gray")
	}
}
