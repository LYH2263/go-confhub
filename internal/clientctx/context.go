// Package clientctx 从 HTTP 头解析灰度客户端身份。
package clientctx

import (
	"net/http"
	"strings"

	"github.com/LYH2263/go-confhub/internal/meta"
)

const (
	HeaderClientID = "X-Confhub-Client-Id"
	HeaderTags     = "X-Confhub-Tags"
	QueryClientID  = "client_id"
	QueryTags      = "tags"
)

func ParseTags(s string) map[string]string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	out := make(map[string]string)
	parts := strings.Split(s, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		k, v, ok := strings.Cut(p, "=")
		if !ok {
			k, v, ok = strings.Cut(p, ":")
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if k == "" {
			continue
		}
		out[k] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func FromRequest(r *http.Request) meta.ClientContext {
	id := r.Header.Get(HeaderClientID)
	if id == "" {
		id = r.URL.Query().Get(QueryClientID)
	}
	tags := ParseTags(r.Header.Get(HeaderTags))
	if tags == nil {
		tags = ParseTags(r.URL.Query().Get(QueryTags))
	}
	return meta.ClientContext{ID: id, Tags: tags}
}

func ApplyTo(h http.Header, c meta.ClientContext) {
	if c.ID != "" {
		h.Set(HeaderClientID, c.ID)
	}
	if len(c.Tags) == 0 {
		return
	}
	parts := make([]string, 0, len(c.Tags))
	for k, v := range c.Tags {
		parts = append(parts, k+"="+v)
	}
	h.Set(HeaderTags, strings.Join(parts, ","))
}

func ParseActor(r *http.Request) string {
	if a := r.Header.Get("X-Confhub-Actor"); a != "" {
		return a
	}
	return r.Header.Get("X-Actor")
}
