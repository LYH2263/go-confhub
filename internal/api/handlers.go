package api

import (
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/LYH2263/go-confhub"
	"github.com/LYH2263/go-confhub/internal/clientctx"
	"github.com/LYH2263/go-confhub/internal/meta"
)

type createNSBody struct {
	ID       string `json:"id"`
	Owner    string `json:"owner"`
	MaxKeys  int    `json:"max_keys"`
	MaxBytes int64  `json:"max_bytes"`
}

func (s *Server) handleListNS(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"namespaces": s.hub.ListNS()})
}

func (s *Server) handleCreateNS(w http.ResponseWriter, r *http.Request) {
	var body createNSBody
	if err := readJSON(r, &body); err != nil {
		writeErr(w, err)
		return
	}
	n, err := s.hub.CreateNSSpec(confhub.NSSpec{
		ID: body.ID, Owner: body.Owner, MaxKeys: body.MaxKeys, MaxBytes: body.MaxBytes,
		Actor: clientctx.ParseActor(r),
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, n)
}

func (s *Server) handleGetNS(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("ns")
	n, err := s.hub.GetNS(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	keys, _ := s.hub.ListKeys(id)
	k, b, _ := s.hub.QuotaUsage(id)
	writeJSON(w, http.StatusOK, map[string]any{
		"namespace": n, "keys": keys, "usage_keys": k, "usage_bytes": b,
	})
}

func (s *Server) handleDeleteNS(w http.ResponseWriter, r *http.Request) {
	if err := s.hub.DeleteNSActor(r.PathValue("ns"), clientctx.ParseActor(r)); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleListKeys(w http.ResponseWriter, r *http.Request) {
	nsID := r.PathValue("ns")
	keys, err := s.hub.ListKeys(nsID)
	if err != nil {
		writeErr(w, err)
		return
	}
	type item struct {
		Key    string `json:"key"`
		Head   int64  `json:"head"`
		Stable int64  `json:"stable"`
		Bytes  int    `json:"bytes"`
	}
	out := make([]item, 0, len(keys))
	for _, k := range keys {
		e, err := s.hub.GetEntry(nsID, k)
		if err != nil {
			continue
		}
		sz := 0
		if v := e.HeadMeta(); v != nil {
			sz = len(v.Payload)
		}
		out = append(out, item{Key: k, Head: e.Head, Stable: e.Stable, Bytes: sz})
	}
	writeJSON(w, http.StatusOK, map[string]any{"keys": out})
}

func (s *Server) handleGetKey(w http.ResponseWriter, r *http.Request) {
	nsID := r.PathValue("ns")
	key := r.PathValue("key")
	c := clientctx.FromRequest(r)
	v, err := s.hub.Get(nsID, key, c)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

type putBody struct {
	Payload    string         `json:"payload"`
	PayloadB64 string         `json:"payload_b64"`
	Author     string         `json:"author"`
	Algo       string         `json:"algo"`
	Signature  string         `json:"signature"`
	KeyID      string         `json:"key_id"`
	Gray       *meta.GrayRule `json:"gray"`
	BindRev    int64          `json:"bind_rev"`
}

func (s *Server) handlePutKey(w http.ResponseWriter, r *http.Request) {
	nsID := r.PathValue("ns")
	key := r.PathValue("key")
	var body putBody
	if err := readJSON(r, &body); err != nil {
		writeErr(w, err)
		return
	}
	payload := []byte(body.Payload)
	if body.PayloadB64 != "" {
		b, err := base64.StdEncoding.DecodeString(body.PayloadB64)
		if err != nil {
			writeErr(w, err)
			return
		}
		payload = b
	}
	var sig []byte
	if body.Signature != "" {
		b, err := base64.StdEncoding.DecodeString(body.Signature)
		if err != nil {
			b, err = parseHexOrRaw(body.Signature)
			if err != nil {
				writeErr(w, err)
				return
			}
		}
		sig = b
	}
	rev, err := s.hub.Put(nsID, key, payload, confhub.PutOptions{
		Author:    body.Author,
		Actor:     clientctx.ParseActor(r),
		Algo:      body.Algo,
		Signature: sig,
		KeyID:     body.KeyID,
		Gray:      body.Gray,
		BindRev:   body.BindRev,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rev": rev, "ns": nsID, "key": key})
}

func (s *Server) handleDeleteKey(w http.ResponseWriter, r *http.Request) {
	if err := s.hub.DeleteKey(r.PathValue("ns"), r.PathValue("key"), clientctx.ParseActor(r)); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type rollbackBody struct {
	Rev int64 `json:"rev"`
}

func (s *Server) handleRollback(w http.ResponseWriter, r *http.Request) {
	var body rollbackBody
	if err := readJSON(r, &body); err != nil {
		writeErr(w, err)
		return
	}
	if err := s.hub.RollbackActor(r.PathValue("ns"), r.PathValue("key"), body.Rev, clientctx.ParseActor(r)); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "head": body.Rev})
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	items, err := s.hub.History(r.PathValue("ns"), r.PathValue("key"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"history": items})
}

func (s *Server) handleAudit(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	recs := s.hub.AuditQuery(r.PathValue("ns"), r.URL.Query().Get("key"), limit)
	writeJSON(w, http.StatusOK, map[string]any{"audit": recs})
}

func (s *Server) handleAuditAll(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	writeJSON(w, http.StatusOK, map[string]any{"audit": s.hub.AuditRecent(limit)})
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	raw, err := s.hub.ExportBytes()
	if err != nil {
		writeErr(w, err)
		return
	}
	if r.URL.Query().Get("format") == "json" {
		b, err := s.hub.ExportSnapshot()
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, b)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
}

func (s *Server) handleImport(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 32*1024)
	for {
		n, err := r.Body.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			break
		}
	}
	if err := s.hub.ImportBytes(buf); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func parseHexOrRaw(s string) ([]byte, error) {
	if b, err := hex.DecodeString(s); err == nil {
		return b, nil
	}
	return []byte(s), nil
}
