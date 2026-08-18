package confhub

import (
	"github.com/LYH2263/go-confhub/internal/meta"
	"github.com/LYH2263/go-confhub/internal/snapshot"
)

func (h *Hub) ExportSnapshot() (*snapshot.Blob, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.checkOpen(); err != nil {
		return nil, err
	}
	return snapshot.Build(snapshot.Source{
		Namespaces: h.ns.List(),
		Entries:    h.store.AllEntries(),
		Audit:      h.audit.All(),
		Grants:     h.ns.ACL().Export(),
		Usage:      h.quota.Export(),
		Now:        h.clk.Now(),
	}), nil
}

func (h *Hub) ImportSnapshot(b *snapshot.Blob) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.checkOpen(); err != nil {
		return err
	}
	if b == nil {
		return nil
	}
	h.ns.ReplaceAll(b.Namespaces)
	if b.Grants != nil {
		h.ns.ACL().Import(b.Grants)
	}
	if err := h.store.ReplaceAll(b.Entries); err != nil {
		return err
	}
	if b.Usage != nil {
		h.quota.ReplaceAll(b.Usage)
	} else {
		h.recalcQuotaLocked()
	}
	if b.Audit != nil {
		h.audit.ReplaceAll(b.Audit)
	}
	h.audit.OK("", "", meta.KindImport, "", 0, len(b.Entries), "")
	h.maybePersistLocked()
	return nil
}

func (h *Hub) recalcQuotaLocked() {
	for _, n := range h.ns.List() {
		h.quota.DeleteNS(n.ID)
	}
	type acc struct {
		keys  int
		bytes int64
	}
	m := map[string]*acc{}
	for _, e := range h.store.AllEntries() {
		a := m[e.NS]
		if a == nil {
			a = &acc{}
			m[e.NS] = a
		}
		a.keys++
		for i := range e.Versions {
			a.bytes += int64(len(e.Versions[i].Payload))
		}
	}
	for nsID, a := range m {
		h.quota.Recalc(nsID, a.keys, a.bytes)
	}
}

func (h *Hub) ExportBytes() ([]byte, error) {
	b, err := h.ExportSnapshot()
	if err != nil {
		return nil, err
	}
	return snapshot.Encode(b)
}

func (h *Hub) ImportBytes(raw []byte) error {
	b, err := snapshot.Decode(raw)
	if err != nil {
		return err
	}
	return h.ImportSnapshot(b)
}
