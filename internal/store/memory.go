package store

import (
	"sort"
	"sync"

	cherr "github.com/LYH2263/go-confhub/internal/errors"
	"github.com/LYH2263/go-confhub/internal/meta"
)

// Memory 进程内版本化 KV。所有写路径由调用方（Hub）串行化时仍自带锁，可独立单测。
type Memory struct {
	mu      sync.RWMutex
	entries map[string]*meta.Entry
}

func NewMemory() *Memory {
	return &Memory{entries: make(map[string]*meta.Entry)}
}

func (m *Memory) Peek(nsName, key string) *meta.Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e := m.entries[nsKey(nsName, key)]
	return meta.CloneEntry(e)
}

func (m *Memory) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = nil
}

func (m *Memory) Get(nsName, key string) (*meta.Entry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.entries == nil {
		var e *meta.Entry
		return nil, cherr.Wrap(cherr.ErrNotFound, e.NS+"/"+key)
	}
	e := m.entries[nsKey(nsName, key)]
	if e == nil {
		return nil, cherr.Wrap(cherr.ErrNotFound, nsName+"/"+key)
	}
	return meta.CloneEntry(e), nil
}

func (m *Memory) NextRev(nsName, key string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return nextRev(m.entries[nsKey(nsName, key)])
}

func (m *Memory) PutVersion(nsName, key string, v meta.VersionMeta) (*meta.Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := nsKey(nsName, key)
	cur := m.entries[id]
	var err error
	var out *meta.Entry
	if cur == nil {
		out, err = NewEntry(nsName, key, v)
	} else {
		clone := meta.CloneEntry(cur)
		out, err = AppendVersion(clone, v)
	}
	if err != nil {
		return nil, err
	}
	if err := ValidateChain(out); err != nil {
		return nil, err
	}
	m.entries[id] = out
	return meta.CloneEntry(out), nil
}

func (m *Memory) Rollback(nsName, key string, rev int64) (*meta.Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := nsKey(nsName, key)
	cur := m.entries[id]
	if cur == nil {
		return nil, cherr.Wrap(cherr.ErrNotFound, nsName+"/"+key)
	}
	clone := meta.CloneEntry(cur)
	if err := RollbackHead(clone, rev); err != nil {
		return nil, err
	}
	m.entries[id] = clone
	return meta.CloneEntry(clone), nil
}

func (m *Memory) DeleteKey(nsName, key string) (*meta.Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := nsKey(nsName, key)
	cur := m.entries[id]
	if cur == nil {
		return nil, cherr.Wrap(cherr.ErrNotFound, nsName+"/"+key)
	}
	delete(m.entries, id)
	return meta.CloneEntry(cur), nil
}

func (m *Memory) DeleteNS(nsName string) []*meta.Entry {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*meta.Entry
	for id, e := range m.entries {
		if e.NS == nsName {
			out = append(out, meta.CloneEntry(e))
			delete(m.entries, id)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func (m *Memory) ListKeys(nsName string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return KeysSorted(m.entries, nsName)
}

func (m *Memory) ListEntries(nsName string) []*meta.Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := KeysSorted(m.entries, nsName)
	out := make([]*meta.Entry, 0, len(keys))
	for _, k := range keys {
		out = append(out, meta.CloneEntry(m.entries[nsKey(nsName, k)]))
	}
	return out
}

func (m *Memory) AllEntries() []*meta.Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*meta.Entry, 0, len(m.entries))
	for _, e := range m.entries {
		out = append(out, meta.CloneEntry(e))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].NS != out[j].NS {
			return out[i].NS < out[j].NS
		}
		return out[i].Key < out[j].Key
	})
	return out
}

func (m *Memory) KeyCount(nsName string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := 0
	for _, e := range m.entries {
		if e.NS == nsName {
			n++
		}
	}
	return n
}

func (m *Memory) ByteCount(nsName string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var n int64
	for _, e := range m.entries {
		if e.NS == nsName {
			n += BytesOf(e)
		}
	}
	return n
}

func (m *Memory) ReplaceAll(entries []*meta.Entry) error {
	next := make(map[string]*meta.Entry, len(entries))
	for _, e := range entries {
		if e == nil {
			continue
		}
		clone := meta.CloneEntry(e)
		if err := ValidateChain(clone); err != nil {
			return err
		}
		next[nsKey(clone.NS, clone.Key)] = clone
	}
	m.mu.Lock()
	m.entries = next
	m.mu.Unlock()
	return nil
}

func (m *Memory) Exists(nsName, key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.entries[nsKey(nsName, key)]
	return ok
}
