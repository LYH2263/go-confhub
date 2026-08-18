package confhub

import (
	cherr "github.com/LYH2263/go-confhub/internal/errors"
	"github.com/LYH2263/go-confhub/internal/gray"
	"github.com/LYH2263/go-confhub/internal/meta"
	"github.com/LYH2263/go-confhub/internal/ns"
	"github.com/LYH2263/go-confhub/internal/quota"
	"github.com/LYH2263/go-confhub/internal/sign"
	"github.com/LYH2263/go-confhub/internal/watch"
)

func (h *Hub) SignPayload(nsName, key, keyID, algo string, rev int64, payload []byte) ([]byte, error) {
	return h.keys.Sign(nsName, key, keyID, algo, rev, payload)
}

func (h *Hub) Put(nsName, key string, payload []byte, opts PutOptions) (int64, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.checkOpen(); err != nil {
		return 0, err
	}
	if err := ns.ValidKey(key); err != nil {
		return 0, err
	}
	if len(payload) == 0 {
		return 0, cherr.ErrPayloadEmpty
	}
	nmeta, err := h.ns.Get(nsName)
	if err != nil {
		return 0, err
	}
	if opts.Actor != "" {
		if err := h.ns.ACL().Check(nsName, opts.Actor, ns.RoleWriter); err != nil {
			h.audit.Fail(nsName, key, meta.KindPublish, opts.Actor, err.Error())
			return 0, err
		}
	}
	g, err := gray.Normalize(opts.Gray)
	if err != nil {
		h.audit.Fail(nsName, key, meta.KindPublish, opts.Actor, err.Error())
		return 0, err
	}
	next := h.store.NextRev(nsName, key)
	if opts.BindRev > 0 && opts.BindRev != next {
		err := cherr.Wrap(cherr.ErrConflict, "bind_rev")
		h.audit.Fail(nsName, key, meta.KindPublish, opts.Actor, err.Error())
		return 0, err
	}
	signRev := int64(0)
	if opts.BindRev > 0 {
		signRev = opts.BindRev
	}
	needSign := h.requireSign || opts.Algo != "" || len(opts.Signature) > 0

	author := opts.Author
	if author == "" {
		author = opts.Actor
	}
	algo := sign.NormalizeAlgo(opts.Algo)
	v := meta.VersionMeta{
		Rev:     next,
		Payload: meta.CloneBytes(payload),
		Sig:     meta.CloneBytes(opts.Signature),
		Algo:    algo,
		Author:  author,
		Ts:      h.clk.Now(),
		Gray:    g,
	}
	newKey := !h.store.Exists(nsName, key)
	entry, err := h.store.PutVersion(nsName, key, v)
	if err != nil {
		h.audit.Fail(nsName, key, meta.KindPublish, opts.Actor, err.Error())
		return 0, err
	}

	if needSign {
		if len(opts.Signature) == 0 {
			h.audit.Fail(nsName, key, meta.KindPublish, opts.Actor, cherr.ErrSignRequired.Error())
			return 0, cherr.ErrSignRequired
		}
		if algo == "" {
			algo = sign.AlgoHMAC
		}
		if err := h.keys.Verify(nsName, key, opts.KeyID, algo, signRev, payload, opts.Signature); err != nil {
			h.audit.Fail(nsName, key, meta.KindPublish, opts.Actor, err.Error())
			_ = h.store.DropLast(nsName, key)
			return 0, err
		}
	}

	if err := h.quota.Check(quota.CheckInput{
		NS:       nmeta,
		NewKey:   newKey,
		AddBytes: int64(len(payload)),
	}); err != nil {
		h.audit.Fail(nsName, key, meta.KindPublish, opts.Actor, err.Error())
		_ = h.store.DropLast(nsName, key)
		return 0, err
	}

	addKeys := 0
	if newKey {
		addKeys = 1
	}
	h.quota.Commit(nsName, addKeys, int64(len(payload)))
	h.audit.OK(nsName, key, meta.KindPublish, author, entry.Head, len(payload), "")
	h.watch.Publish(watch.NewEvent(nsName, key, meta.KindPublish, entry.Head))
	h.maybePersistLocked()
	return entry.Head, nil
}
