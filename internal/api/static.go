package api

import (
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if isAPI(r.URL.Path) {
		http.NotFound(w, r)
		return
	}
	if s.opts.WebFS != nil {
		s.serveFS(w, r, s.opts.WebFS)
		return
	}
	dir := lookupWeb(s.opts.WebDir)
	fsys := os.DirFS(dir)
	s.serveFS(w, r, fsys)
}

func (s *Server) serveFS(w http.ResponseWriter, r *http.Request, fsys fs.FS) {
	p := r.URL.Path
	if p == "/" || p == "" {
		p = "index.html"
	} else {
		p = strings.TrimPrefix(p, "/")
	}
	p = path.Clean(p)
	if strings.Contains(p, "..") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	f, err := fsys.Open(p)
	if err != nil {
		if p != "index.html" {
			if _, err2 := fsys.Open("index.html"); err2 == nil {
				http.ServeFileFS(w, r, fsys, "index.html")
				return
			}
		}
		http.NotFound(w, r)
		return
	}
	_ = f.Close()
	http.ServeFileFS(w, r, fsys, p)
}
