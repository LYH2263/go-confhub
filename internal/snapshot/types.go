package snapshot

import (
	"time"

	"github.com/LYH2263/go-confhub/internal/meta"
	"github.com/LYH2263/go-confhub/internal/quota"
)

const Magic = "CHB1"
const Format = 1

type Blob struct {
	Header     meta.SnapshotHeader          `json:"header"`
	Namespaces []meta.Namespace             `json:"namespaces"`
	Entries    []*meta.Entry                `json:"entries"`
	Audit      []meta.AuditRecord           `json:"audit,omitempty"`
	Grants     map[string]map[string]string `json:"grants,omitempty"`
	Usage      map[string]quota.Usage       `json:"usage,omitempty"`
}

func NewBlob(now time.Time) *Blob {
	return &Blob{
		Header: meta.SnapshotHeader{
			Magic:     Magic,
			Format:    Format,
			CreatedAt: now.UTC(),
		},
		Grants: make(map[string]map[string]string),
		Usage:  make(map[string]quota.Usage),
	}
}

func (b *Blob) Tally() {
	b.Header.NSCount = len(b.Namespaces)
	b.Header.KeyCount = len(b.Entries)
	n := 0
	for _, e := range b.Entries {
		if e != nil {
			n += len(e.Versions)
		}
	}
	b.Header.VerCount = n
}
