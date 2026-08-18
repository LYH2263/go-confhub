package quota

import (
	cherr "github.com/LYH2263/go-confhub/internal/errors"
	"github.com/LYH2263/go-confhub/internal/meta"
)

type Limiter struct {
	ctr *Counter
}

func NewLimiter() *Limiter {
	return &Limiter{ctr: NewCounter()}
}

func (l *Limiter) Counter() *Counter { return l.ctr }

type CheckInput struct {
	NS       meta.Namespace
	NewKey   bool
	AddBytes int64
}

// Check 只判断，不记账。失败时调用方不得写入版本。
func (l *Limiter) Check(in CheckInput) error {
	if in.NS.ID == "" {
		return cherr.Wrap(cherr.ErrInvalidArg, "quota ns")
	}
	u := l.ctr.Get(in.NS.ID)
	keys := u.Keys
	if in.NewKey {
		keys++
	}
	if in.NS.MaxKeys > 0 && keys > in.NS.MaxKeys {
		return cherr.Wrap(cherr.ErrQuotaKeys, in.NS.ID)
	}
	bytes := u.Bytes + in.AddBytes
	if in.AddBytes < 0 {
		bytes = u.Bytes + in.AddBytes
	}
	if in.NS.MaxBytes > 0 && bytes > in.NS.MaxBytes {
		return cherr.Wrap(cherr.ErrQuotaBytes, in.NS.ID)
	}
	return nil
}

func (l *Limiter) Commit(ns string, keys int, bytes int64) Usage {
	return l.ctr.Apply(ns, keys, bytes)
}

func (l *Limiter) Usage(ns string) Usage {
	return l.ctr.Get(ns)
}

func (l *Limiter) Recalc(ns string, keys int, bytes int64) {
	l.ctr.Set(ns, Usage{Keys: keys, Bytes: bytes})
}

func (l *Limiter) DeleteNS(ns string) {
	l.ctr.DeleteNS(ns)
}

func (l *Limiter) ReplaceAll(m map[string]Usage) {
	l.ctr.Import(m)
}

func (l *Limiter) Export() map[string]Usage {
	return l.ctr.Export()
}
