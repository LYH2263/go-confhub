package ns

import (
	"sort"
	"sync"
	"time"

	"github.com/LYH2263/go-confhub/internal/clock"
	cherr "github.com/LYH2263/go-confhub/internal/errors"
	"github.com/LYH2263/go-confhub/internal/meta"
)

const (
	DefaultMaxKeys  = 1024
	DefaultMaxBytes = 32 << 20
)

type Registry struct {
	mu    sync.RWMutex
	items map[string]meta.Namespace
	acl   *ACL
	clk   clock.Clock
}

func NewRegistry(clk clock.Clock, acl *ACL) *Registry {
	if clk == nil {
		clk = clock.Real{}
	}
	if acl == nil {
		acl = NewACL()
	}
	return &Registry{
		items: make(map[string]meta.Namespace),
		acl:   acl,
		clk:   clk,
	}
}

func (r *Registry) ACL() *ACL { return r.acl }

func (r *Registry) Create(id, owner string, maxKeys int, maxBytes int64) (meta.Namespace, error) {
	if err := ValidNSID(id); err != nil {
		return meta.Namespace{}, err
	}
	if err := ValidOwner(owner); err != nil {
		return meta.Namespace{}, err
	}
	if maxKeys < 0 || maxBytes < 0 {
		return meta.Namespace{}, cherr.Wrap(cherr.ErrInvalidArg, "quota")
	}
	if maxKeys == 0 {
		maxKeys = DefaultMaxKeys
	}
	if maxBytes == 0 {
		maxBytes = DefaultMaxBytes
	}
	ns := meta.Namespace{
		ID:        id,
		Owner:     owner,
		CreatedAt: r.clk.Now(),
		MaxKeys:   maxKeys,
		MaxBytes:  maxBytes,
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; ok {
		return meta.Namespace{}, cherr.Wrap(cherr.ErrAlreadyExists, id)
	}
	r.items[id] = ns
	r.acl.SetOwner(id, owner)
	return ns, nil
}

func (r *Registry) Get(id string) (meta.Namespace, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ns, ok := r.items[id]
	if !ok {
		return meta.Namespace{}, cherr.Wrap(cherr.ErrNotFound, id)
	}
	return ns, nil
}

func (r *Registry) MustExist(id string) error {
	_, err := r.Get(id)
	return err
}

func (r *Registry) Delete(id string) (meta.Namespace, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ns, ok := r.items[id]
	if !ok {
		return meta.Namespace{}, cherr.Wrap(cherr.ErrNotFound, id)
	}
	delete(r.items, id)
	r.acl.RemoveNS(id)
	return ns, nil
}

func (r *Registry) List() []meta.Namespace {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]meta.Namespace, 0, len(r.items))
	for _, v := range r.items {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (r *Registry) UpdateQuota(id string, maxKeys int, maxBytes int64) (meta.Namespace, error) {
	if maxKeys < 0 || maxBytes < 0 {
		return meta.Namespace{}, cherr.Wrap(cherr.ErrInvalidArg, "quota")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	ns, ok := r.items[id]
	if !ok {
		return meta.Namespace{}, cherr.Wrap(cherr.ErrNotFound, id)
	}
	if maxKeys > 0 {
		ns.MaxKeys = maxKeys
	}
	if maxBytes > 0 {
		ns.MaxBytes = maxBytes
	}
	r.items[id] = ns
	return ns, nil
}

func (r *Registry) ReplaceAll(items []meta.Namespace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = make(map[string]meta.Namespace, len(items))
	for _, n := range items {
		r.items[n.ID] = n
		r.acl.SetOwner(n.ID, n.Owner)
	}
}

func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items)
}

func (r *Registry) Exists(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.items[id]
	return ok
}

func (r *Registry) CreatedAt(id string) time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.items[id].CreatedAt
}
