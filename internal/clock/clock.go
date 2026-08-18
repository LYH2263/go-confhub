// Package clock 提供可注入时钟，便于审计时间戳与测试冻结。
package clock

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type Real struct{}

func (Real) Now() time.Time { return time.Now().UTC() }

// Fake 单调可推进的测试时钟。
type Fake struct {
	mu sync.Mutex
	t  time.Time
}

func NewFake(t time.Time) *Fake {
	if t.IsZero() {
		t = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	return &Fake{t: t.UTC()}
}

func (f *Fake) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.t
}

func (f *Fake) Set(t time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.t = t.UTC()
}

func (f *Fake) Advance(d time.Duration) time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.t = f.t.Add(d)
	return f.t
}
