package audit

import (
	"time"

	"github.com/LYH2263/go-confhub/internal/meta"
)

type Record = meta.AuditRecord

type Filter struct {
	NS    string
	Key   string
	Kind  string
	Actor string
	Since time.Time
	Until time.Time
	OK    *bool
	Limit int
}

func (f Filter) Match(r Record) bool {
	if f.NS != "" && r.NS != f.NS {
		return false
	}
	if f.Key != "" && r.Key != f.Key {
		return false
	}
	if f.Kind != "" && r.Kind != f.Kind {
		return false
	}
	if f.Actor != "" && r.Actor != f.Actor {
		return false
	}
	if !f.Since.IsZero() && r.Ts.Before(f.Since) {
		return false
	}
	if !f.Until.IsZero() && !r.Ts.Before(f.Until) && !r.Ts.Equal(f.Until) {
		return false
	}
	if f.OK != nil && r.OK != *f.OK {
		return false
	}
	return true
}

func CloneRecord(r Record) Record {
	return r
}
