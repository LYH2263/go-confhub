package confhub

import (
	"github.com/LYH2263/go-confhub/internal/clock"
	"github.com/LYH2263/go-confhub/internal/sign"
)

type Option func(*Hub)

func WithClock(c clock.Clock) Option {
	return func(h *Hub) {
		if c != nil {
			h.clk = c
		}
	}
}

func WithRequireSign(v bool) Option {
	return func(h *Hub) { h.requireSign = v }
}

func WithAuditCap(n int) Option {
	return func(h *Hub) { h.auditCap = n }
}

func WithPersistPath(p string) Option {
	return func(h *Hub) { h.persistPath = p }
}

func WithHMACSecret(secret []byte) Option {
	return func(h *Hub) {
		_ = h.keys.PutHMAC("", h.keys.DefaultID(), secret)
	}
}

func WithEd25519(pair sign.EdKeyPair) Option {
	return func(h *Hub) {
		_ = h.keys.PutEd("", h.keys.DefaultID(), pair)
	}
}

func WithDefaultKeyID(id string) Option {
	return func(h *Hub) { h.keys.SetDefaultID(id) }
}
