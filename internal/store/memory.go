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

func (m *Memory) Get(nsName, key string) (*meta.Entry, error) {
	e := m.Peek(nsName, key)
	if e == nil {
		return nil, cherr.Wrap(cherr.ErrNotFound, nsName+"/"+key)
	}
	return e, nil
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
		out, err = AppendVersion(cur, v)
	}
	if err != nil {
		return nil, err
	}
	m.entries[id] = out
	return out, nil
}

// DropLast 撤掉刚写入的末尾版本：签名或配额校验失败时，PutVersion 已追加的
// 版本必须回滚，否则脏版本会对读路径（Get/ListKeys）可见。
// 仅剩一条版本时连同整个条目一起删除；否则去掉末尾版本并重算 Head/Stable。
func (m *Memory) DropLast(nsName, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := nsKey(nsName, key)
	cur := m.entries[id]
	if cur == nil || len(cur.Versions) == 0 {
		return nil
	}
	if len(cur.Versions) == 1 {
		delete(m.entries, id)
		return nil
	}
	cur.Versions[len(cur.Versions)-1] = meta.VersionMeta{}
	cur.Versions = cur.Versions[:len(cur.Versions)-1]
	last := cur.Versions[len(cur.Versions)-1]
	cur.Head = last.Rev
	cur.Stable = lastStableAtOrBefore(cur, last.Rev)
	return nil
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
