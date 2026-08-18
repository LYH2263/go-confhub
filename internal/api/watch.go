package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/LYH2263/go-confhub/internal/watch"
)

func (s *Server) handleWatch(w http.ResponseWriter, r *http.Request) {
	nsID := r.PathValue("ns")
	key := r.PathValue("key")
	since, _ := strconv.ParseInt(r.URL.Query().Get("since"), 10, 64)
	timeout := queryDuration(r, "timeout", 25*time.Second)
	ev, ok := s.hub.WatchHub().WaitContext(r.Context(), nsID, key, since, timeout)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{"timeout": true})
		return
	}
	writeJSON(w, http.StatusOK, watch.Event(ev))
}
