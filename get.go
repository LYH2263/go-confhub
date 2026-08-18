package confhub

import (
	cherr "github.com/LYH2263/go-confhub/internal/errors"
	"github.com/LYH2263/go-confhub/internal/gray"
	"github.com/LYH2263/go-confhub/internal/meta"
	"github.com/LYH2263/go-confhub/internal/ns"
	"github.com/LYH2263/go-confhub/internal/watch"
)

func (h *Hub) Get(nsName, key string, client ClientContext) (*Value, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.checkOpen(); err != nil {
		return nil, err
	}
	if err := h.ns.MustExist(nsName); err != nil {
		return nil, err
	}
	e, err := h.store.Get(nsName, key)
	if err != nil {
		return nil, err
	}
	rev, hit, ok := gray.Select(e, client)
	if !ok {
		return nil, cherr.Wrap(cherr.ErrNoStable, nsName+"/"+key)
	}
	v := e.ByRev(rev)
	if v == nil {
		return nil, cherr.Wrap(cherr.ErrNotFound, nsName+"/"+key)
	}
	return valueFrom(e, v, hit), nil
}

func (h *Hub) GetRev(nsName, key string, rev int64) (*Value, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.checkOpen(); err != nil {
		return nil, err
	}
	e, err := h.store.Get(nsName, key)
	if err != nil {
		return nil, err
	}
	v := e.ByRev(rev)
	if v == nil {
		return nil, cherr.Wrap(cherr.ErrBadRevision, nsName+"/"+key)
	}
	return valueFrom(e, v, rev == e.Head), nil
}

func (h *Hub) Rollback(nsName, key string, rev int64) error {
	return h.RollbackActor(nsName, key, rev, "")
}

func (h *Hub) RollbackActor(nsName, key string, rev int64, actor string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.checkOpen(); err != nil {
		return err
	}
	if err := h.ns.MustExist(nsName); err != nil {
		return err
	}
	if actor != "" {
		if err := h.ns.ACL().Check(nsName, actor, ns.RoleWriter); err != nil {
			h.audit.Fail(nsName, key, meta.KindRollback, actor, err.Error())
			return err
		}
	}
	e, err := h.store.Rollback(nsName, key, rev)
	if err != nil {
		h.audit.Fail(nsName, key, meta.KindRollback, actor, err.Error())
		return err
	}
	h.audit.OK(nsName, key, meta.KindRollback, actor, e.Head, 0, "")
	h.watch.Publish(watch.NewEvent(nsName, key, meta.KindRollback, e.Head))
	if h.persistPath != "" {
		_ = h.store.SaveJSON(h.persistPath)
	}
	return nil
}

func (h *Hub) DeleteKey(nsName, key, actor string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.checkOpen(); err != nil {
		return err
	}
	if actor != "" {
		if err := h.ns.ACL().Check(nsName, actor, ns.RoleWriter); err != nil {
			return err
		}
	}
	e, err := h.store.DeleteKey(nsName, key)
	if err != nil {
		h.audit.Fail(nsName, key, meta.KindDelete, actor, err.Error())
		return err
	}
	var bytes int64
	for i := range e.Versions {
		bytes += int64(len(e.Versions[i].Payload))
	}
	h.quota.Commit(nsName, -1, -bytes)
	h.audit.OK(nsName, key, meta.KindDelete, actor, e.Head, int(bytes), "")
	h.watch.Publish(watch.NewEvent(nsName, key, meta.KindDelete, e.Head))
	h.maybePersistLocked()
	return nil
}

func (h *Hub) Watch(nsName, key string) (<-chan WatchEvent, func()) {
	return h.watch.Subscribe(nsName, key)
}
