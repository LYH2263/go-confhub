package audit

import (
	"sync"

	"github.com/LYH2263/go-confhub/internal/clock"
)

const DefaultCap = 4096

// Log 环形审计。超过 cap 丢弃最旧记录，seq 仍单调递增。
type Log struct {
	mu   sync.Mutex
	clk  clock.Clock
	cap  int
	seq  int64
	buf  []Record
	head int // 最旧元素下标
	len  int
}

func NewLog(clk clock.Clock, cap int) *Log {
	if clk == nil {
		clk = clock.Real{}
	}
	if cap <= 0 {
		cap = DefaultCap
	}
	return &Log{clk: clk, cap: cap, buf: make([]Record, cap)}
}

func (l *Log) Append(r Record) Record {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.seq++
	r.Seq = l.seq
	if r.Ts.IsZero() {
		r.Ts = l.clk.Now()
	}
	if l.len < l.cap {
		idx := (l.head + l.len) % l.cap
		l.buf[idx] = r
		l.len++
	} else {
		l.buf[l.head] = r
		l.head = (l.head + 1) % l.cap
	}
	return r
}

func (l *Log) Recent(n int) []Record {
	return l.Query(Filter{Limit: n})
}

func (l *Log) Query(f Filter) []Record {
	l.mu.Lock()
	defer l.mu.Unlock()
	if f.Limit <= 0 {
		f.Limit = 100
	}
	out := make([]Record, 0, min(f.Limit, l.len))
	for i := l.len - 1; i >= 0 && len(out) < f.Limit; i-- {
		idx := (l.head + i) % l.cap
		r := l.buf[idx]
		if f.Match(r) {
			out = append(out, r)
		}
	}
	return out
}

func (l *Log) All() []Record {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Record, 0, l.len)
	for i := 0; i < l.len; i++ {
		idx := (l.head + i) % l.cap
		out = append(out, l.buf[idx])
	}
	return out
}

func (l *Log) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.len
}

func (l *Log) Seq() int64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.seq
}

func (l *Log) ReplaceAll(recs []Record) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.head = 0
	l.len = 0
	l.seq = 0
	for _, r := range recs {
		if r.Seq > l.seq {
			l.seq = r.Seq
		}
		if l.len < l.cap {
			l.buf[l.len] = r
			l.len++
		} else {
			l.buf[l.head] = r
			l.head = (l.head + 1) % l.cap
		}
	}
}

func (l *Log) Fail(ns, key, kind, actor, err string) Record {
	return l.Append(Record{
		NS: ns, Key: key, Kind: kind, Actor: actor, OK: false, Err: err,
	})
}

func (l *Log) OK(ns, key, kind, actor string, rev int64, bytes int, detail string) Record {
	r := Record{
		NS: ns, Key: key, Kind: kind, Actor: actor, Rev: rev, Bytes: bytes, Detail: detail, OK: true,
	}
	if r.Ts.IsZero() {
		r.Ts = l.clk.Now()
	}
	return r
}
