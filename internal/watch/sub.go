package watch

import (
	"context"
	"sync"
	"sync/atomic"
)

const defaultBuf = 16

type subID uint64

type subscription struct {
	id     subID
	ns     string
	key    string
	ch     chan Event
	closed atomic.Bool
	ctx    context.Context
}

func (s *subscription) send(ev Event) {
	if s.closed.Load() {
		return
	}
	// ctx 已取消则不再投递：取消后立即到来的事件也不得进入缓冲区，
	// 否则取消与关闭 channel 之间存在竞态，等待方会读到取消后才到达的事件。
	if s.ctx != nil && s.ctx.Err() != nil {
		return
	}
	if !Match(s.ns, s.key, ev) {
		return
	}
	select {
	case s.ch <- ev:
	default:
		// 缓冲区满时丢弃最旧再投递最新，避免阻塞发布路径。
		select {
		case <-s.ch:
		default:
		}
		select {
		case s.ch <- ev:
		default:
		}
	}
}

func (s *subscription) close() {
	if s.closed.Swap(true) {
		return
	}
	close(s.ch)
}

type cancelFn struct {
	once sync.Once
	fn   func()
	done chan struct{}
}

func newCancelFn(fn func()) *cancelFn {
	return &cancelFn{fn: fn, done: make(chan struct{})}
}

// Call 既要拆除订阅，也要唤醒监听 ctx.Done() 的协程退出，避免协程泄漏。
func (c *cancelFn) Call() {
	c.once.Do(func() {
		close(c.done)
		c.fn()
	})
}
