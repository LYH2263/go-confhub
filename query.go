package confhub

import (
	"github.com/LYH2263/go-confhub/internal/audit"
	"github.com/LYH2263/go-confhub/internal/ns"
	"github.com/LYH2263/go-confhub/internal/version"
)

func (h *Hub) ListKeys(nsName string) ([]string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ns.MustExist(nsName); err != nil {
		return nil, err
	}
	return h.store.ListKeys(nsName), nil
}

func (h *Hub) GetEntry(nsName, key string) (*Entry, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ns.MustExist(nsName); err != nil {
		return nil, err
	}
	return h.store.Get(nsName, key)
}

func (h *Hub) History(nsName, key string) ([]version.HistoryItem, error) {
	e, err := h.GetEntry(nsName, key)
	if err != nil {
		return nil, err
	}
	return version.History(e), nil
}

func (h *Hub) Diff(nsName, key string, from, to int64) (version.DiffResult, error) {
	e, err := h.GetEntry(nsName, key)
	if err != nil {
		return version.DiffResult{}, err
	}
	return version.DiffRevs(e, from, to)
}

func (h *Hub) AuditRecent(n int) []AuditRecord {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.audit.Recent(n)
}

func (h *Hub) AuditQuery(nsName, key string, limit int) []AuditRecord {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.audit.Query(audit.Filter{NS: nsName, Key: key, Limit: limit})
}

func (h *Hub) QuotaUsage(nsName string) (keys int, bytes int64, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ns.MustExist(nsName); err != nil {
		return 0, 0, err
	}
	u := h.quota.Usage(nsName)
	return u.Keys, u.Bytes, nil
}

func (h *Hub) Grants(nsName string) ([]ns.Grant, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ns.MustExist(nsName); err != nil {
		return nil, err
	}
	return h.ns.ACL().List(nsName), nil
}
