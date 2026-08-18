package confhub

import (
	"sync"

	"github.com/LYH2263/go-confhub/internal/audit"
	"github.com/LYH2263/go-confhub/internal/clock"
	cherr "github.com/LYH2263/go-confhub/internal/errors"
	"github.com/LYH2263/go-confhub/internal/meta"
	"github.com/LYH2263/go-confhub/internal/ns"
	"github.com/LYH2263/go-confhub/internal/quota"
	"github.com/LYH2263/go-confhub/internal/sign"
	"github.com/LYH2263/go-confhub/internal/store"
	"github.com/LYH2263/go-confhub/internal/watch"
)

// Hub 配置中心核心：命名空间、版本化存储、灰度、Watch、签名、审计、配额。
type Hub struct {
	mu          sync.Mutex
	closed      bool
	clk         clock.Clock
	ns          *ns.Registry
	store       *store.Memory
	watch       *watch.Hub
	audit       *audit.Log
	quota       *quota.Limiter
	keys        *sign.Keyring
	requireSign bool
	auditCap    int
	persistPath string
}

func New(opts ...Option) *Hub {
	h := &Hub{
		clk:   clock.Real{},
		store: store.NewMemory(),
		watch: watch.NewHub(),
		quota: quota.NewLimiter(),
		keys:  sign.NewKeyring(),
	}
	for _, o := range opts {
		o(h)
	}
	acl := ns.NewACL()
	h.ns = ns.NewRegistry(h.clk, acl)
	h.audit = audit.NewLog(h.clk, h.auditCap)
	return h
}

func (h *Hub) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return nil
	}
	h.closed = true
	h.store.Close()
	h.watch.Close()
	return nil
}

func (h *Hub) checkOpen() error {
	if h.closed {
		return cherr.ErrClosed
	}
	return nil
}

func (h *Hub) Keyring() *sign.Keyring { return h.keys }

func (h *Hub) ACL() *ns.ACL { return h.ns.ACL() }

func (h *Hub) WatchHub() *watch.Hub { return h.watch }

func (h *Hub) CreateNS(id, owner string, maxKeys int, maxBytes int64) (Namespace, error) {
	return h.CreateNSSpec(NSSpec{ID: id, Owner: owner, MaxKeys: maxKeys, MaxBytes: maxBytes})
}

func (h *Hub) CreateNSSpec(spec NSSpec) (Namespace, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.checkOpen(); err != nil {
		return Namespace{}, err
	}
	n, err := h.ns.Create(spec.ID, spec.Owner, spec.MaxKeys, spec.MaxBytes)
	if err != nil {
		h.audit.Fail(spec.ID, "", meta.KindCreateNS, spec.Actor, err.Error())
		return Namespace{}, err
	}
	h.audit.OK(n.ID, "", meta.KindCreateNS, spec.Owner, 0, 0, "owner="+n.Owner)
	h.maybePersistLocked()
	return n, nil
}

func (h *Hub) DeleteNS(id string) error {
	return h.DeleteNSActor(id, "")
}

func (h *Hub) DeleteNSActor(id, actor string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.checkOpen(); err != nil {
		return err
	}
	if err := h.ns.MustExist(id); err != nil {
		h.audit.Fail(id, "", meta.KindDeleteNS, actor, err.Error())
		return err
	}
	if actor != "" {
		if err := h.ns.ACL().Check(id, actor, ns.RoleAdmin); err != nil {
			return err
		}
	}
	removed := h.store.DeleteNS(id)
	_, err := h.ns.Delete(id)
	if err != nil {
		h.audit.Fail(id, "", meta.KindDeleteNS, actor, err.Error())
		return err
	}
	h.quota.DeleteNS(id)
	for _, e := range removed {
		h.watch.Publish(watch.NewEvent(e.NS, e.Key, meta.KindDelete, e.Head))
	}
	h.audit.OK(id, "", meta.KindDeleteNS, actor, 0, len(removed), "")
	h.maybePersistLocked()
	return nil
}

func (h *Hub) GetNS(id string) (Namespace, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.ns.Get(id)
}

func (h *Hub) ListNS() []Namespace {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.ns.List()
}

func (h *Hub) Grant(nsID, subject, role, actor string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ns.MustExist(nsID); err != nil {
		return err
	}
	if actor != "" {
		if err := h.ns.ACL().Check(nsID, actor, ns.RoleAdmin); err != nil {
			return err
		}
	}
	r, err := ns.ParseRole(role)
	if err != nil {
		return err
	}
	return h.ns.ACL().Grant(nsID, subject, r)
}

func (h *Hub) maybePersistLocked() {
	if h.persistPath == "" {
		return
	}
	_ = h.store.SaveJSON(h.persistPath)
}
