package api

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LYH2263/go-confhub"
)

type Options struct {
	WebDir    string
	WebFS     fs.FS
	AllowCORS bool
}

type Server struct {
	hub  *confhub.Hub
	mux  *http.ServeMux
	opts Options
}

func New(hub *confhub.Hub, opts Options) *Server {
	s := &Server{hub: hub, mux: http.NewServeMux(), opts: opts}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/ns", s.handleListNS)
	s.mux.HandleFunc("POST /api/ns", s.handleCreateNS)
	s.mux.HandleFunc("GET /api/ns/{ns}", s.handleGetNS)
	s.mux.HandleFunc("DELETE /api/ns/{ns}", s.handleDeleteNS)
	s.mux.HandleFunc("GET /api/ns/{ns}/keys", s.handleListKeys)
	s.mux.HandleFunc("GET /api/ns/{ns}/keys/{key}", s.handleGetKey)
	s.mux.HandleFunc("PUT /api/ns/{ns}/keys/{key}", s.handlePutKey)
	s.mux.HandleFunc("DELETE /api/ns/{ns}/keys/{key}", s.handleDeleteKey)
	s.mux.HandleFunc("POST /api/ns/{ns}/keys/{key}/rollback", s.handleRollback)
	s.mux.HandleFunc("GET /api/ns/{ns}/keys/{key}/watch", s.handleWatch)
	s.mux.HandleFunc("GET /api/ns/{ns}/keys/{key}/history", s.handleHistory)
	s.mux.HandleFunc("GET /api/ns/{ns}/audit", s.handleAudit)
	s.mux.HandleFunc("GET /api/audit", s.handleAuditAll)
	s.mux.HandleFunc("GET /api/snapshot", s.handleExport)
	s.mux.HandleFunc("POST /api/snapshot", s.handleImport)
	s.mux.HandleFunc("GET /", s.handleStatic)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "ts": time.Now().UTC()})
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.opts.AllowCORS {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Confhub-Client-Id, X-Confhub-Tags, X-Confhub-Actor")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	s.mux.ServeHTTP(w, r)
}

func lookupWeb(dir string) string {
	if dir == "" {
		dir = "web"
	}
	if st, err := os.Stat(dir); err == nil && st.IsDir() {
		return dir
	}
	exe, err := os.Executable()
	if err == nil {
		cand := filepath.Join(filepath.Dir(exe), "web")
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			return cand
		}
		cand = filepath.Join(filepath.Dir(exe), "..", "web")
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			return cand
		}
	}
	return dir
}

func isAPI(path string) bool {
	return strings.HasPrefix(path, "/api/") || path == "/api"
}
