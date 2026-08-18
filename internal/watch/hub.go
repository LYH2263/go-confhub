package watch

import (
	"sync"
	"sync/atomic"
	"time"
)

// Hub 管理 (ns,key) 订阅。取消后不得再投递（channel 关闭且从集合移除）。
type Hub struct {
	mu      sync.Mutex
	next    atomic.Uint64
	subs    map[subID]*subscription
	last    map[string]Event
	bufSize int
}

func NewHub() *Hub {
	return &Hub{
		subs:    make(map[subID]*subscription),
		last:    make(map[string]Event),
		bufSize: defaultBuf,
	}
}

func lastKey(ns, key string) string { return ns + "\x00" + key }

func (h *Hub) Subscribe(ns, key string) (<-chan Event, func()) {
	return h.SubscribeBuf(ns, key, h.bufSize)
}

func (h *Hub) SubscribeBuf(ns, key string, buf int) (<-chan Event, func()) {
	if buf < 1 {
		buf = 1
	}
	id := subID(h.next.Add(1))
	s := &subscription{
		id:  id,
		ns:  ns,
		key: key,
		ch:  make(chan Event, buf),
	}
	h.mu.Lock()
	h.subs[id] = s
	h.mu.Unlock()
	c := &cancelFn{fn: func() { h.unsubscribe(id) }}
	return s.ch, c.Call
}

func (h *Hub) unsubscribe(id subID) {
	h.mu.Lock()
	s, ok := h.subs[id]
	if ok {
		delete(h.subs, id)
	}
	h.mu.Unlock()
	if ok {
		s.close()
	}
}

func (h *Hub) Publish(ev Event) {
	h.mu.Lock()
	h.last[lastKey(ev.NS, ev.Key)] = ev
	list := make([]*subscription, 0, len(h.subs))
	for _, s := range h.subs {
		list = append(list, s)
	}
	h.mu.Unlock()
	for _, s := range list {
		s.send(ev)
	}
}

func (h *Hub) Last(ns, key string) (Event, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	ev, ok := h.last[lastKey(ns, key)]
	return ev, ok
}

// Wait 长轮询：若已有 rev>since 的事件则立即返回，否则等到超时。
func (h *Hub) Wait(ns, key string, since int64, timeout time.Duration) (Event, bool) {
	if last, ok := h.Last(ns, key); ok && last.Rev > since {
		return last, true
	}
	ch, cancel := h.Subscribe(ns, key)
	defer cancel()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				return Event{}, false
			}
			if ev.Rev > since {
				return ev, true
			}
		case <-timer.C:
			if last, ok := h.Last(ns, key); ok && last.Rev > since {
				return last, true
			}
			return Event{}, false
		}
	}
}

func (h *Hub) Count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.subs)
}

func (h *Hub) Close() {
	h.mu.Lock()
	list := make([]*subscription, 0, len(h.subs))
	for id, s := range h.subs {
		list = append(list, s)
		delete(h.subs, id)
	}
	h.mu.Unlock()
	for _, s := range list {
		s.close()
	}
}
