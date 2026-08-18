package watch

import (
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
}

func (s *subscription) send(ev Event) {
	if s.closed.Load() {
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
}

func (c *cancelFn) Call() {
	c.once.Do(c.fn)
}
