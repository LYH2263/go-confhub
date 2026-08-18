package version

import (
	cherr "github.com/LYH2263/go-confhub/internal/errors"
	"github.com/LYH2263/go-confhub/internal/meta"
)

type HistoryItem struct {
	Rev    int64  `json:"rev"`
	Author string `json:"author,omitempty"`
	Bytes  int    `json:"bytes"`
	Gray   bool   `json:"gray"`
	Head   bool   `json:"head"`
	Stable bool   `json:"stable"`
	Algo   string `json:"algo,omitempty"`
}

func History(e *meta.Entry) []HistoryItem {
	if e == nil {
		return nil
	}
	out := make([]HistoryItem, 0, len(e.Versions))
	for i := range e.Versions {
		v := e.Versions[i]
		out = append(out, HistoryItem{
			Rev:    v.Rev,
			Author: v.Author,
			Bytes:  len(v.Payload),
			Gray:   v.Gray != nil && !v.Gray.IsZero() && !v.Gray.IsFullRollout(),
			Head:   v.Rev == e.Head,
			Stable: v.Rev == e.Stable,
			Algo:   v.Algo,
		})
	}
	return out
}

func DiffRevs(e *meta.Entry, from, to int64) (DiffResult, error) {
	if e == nil {
		return DiffResult{}, cherr.ErrNotFound
	}
	a := e.ByRev(from)
	b := e.ByRev(to)
	if a == nil || b == nil {
		return DiffResult{}, cherr.Wrap(cherr.ErrBadRevision, Format(from)+".."+Format(to))
	}
	return Diff(a.Payload, b.Payload), nil
}

func CanRollbackTo(e *meta.Entry, rev int64) bool {
	if e == nil {
		return false
	}
	return e.ByRev(rev) != nil
}

func HeadRev(e *meta.Entry) int64 {
	if e == nil {
		return 0
	}
	return e.Head
}

func PrevRev(rev int64) int64 {
	if rev <= 1 {
		return 0
	}
	return rev - 1
}
