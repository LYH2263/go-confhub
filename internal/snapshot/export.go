package snapshot

import (
	"encoding/json"
	"os"
	"time"

	"github.com/LYH2263/go-confhub/internal/meta"
	"github.com/LYH2263/go-confhub/internal/quota"
	"github.com/LYH2263/go-confhub/internal/store"
)

type Source struct {
	Namespaces []meta.Namespace
	Entries    []*meta.Entry
	Audit      []meta.AuditRecord
	Grants     map[string]map[string]string
	Usage      map[string]quota.Usage
	Now        time.Time
}

func Build(src Source) *Blob {
	now := src.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	b := NewBlob(now)
	b.Namespaces = append([]meta.Namespace(nil), src.Namespaces...)
	b.Entries = make([]*meta.Entry, 0, len(src.Entries))
	for _, e := range src.Entries {
		b.Entries = append(b.Entries, meta.CloneEntry(e))
	}
	b.Audit = append([]meta.AuditRecord(nil), src.Audit...)
	b.Grants = src.Grants
	b.Usage = src.Usage
	b.Tally()
	return b
}

func MarshalJSON(b *Blob) ([]byte, error) {
	b.Tally()
	return json.MarshalIndent(b, "", "  ")
}

func WriteFile(path string, b *Blob, jsonFmt bool) error {
	var raw []byte
	var err error
	if jsonFmt {
		raw, err = MarshalJSON(b)
	} else {
		raw, err = Encode(b)
	}
	if err != nil {
		return err
	}
	return store.AtomicWriteFile(path, raw)
}

func ReadFile(path string) (*Blob, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Decode(raw)
}
