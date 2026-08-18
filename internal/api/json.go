package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	cherr "github.com/LYH2263/go-confhub/internal/errors"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}

type errBody struct {
	Error string `json:"error"`
}

func writeErr(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	switch {
	case errors.Is(err, cherr.ErrNotFound), errors.Is(err, cherr.ErrNoStable):
		status = http.StatusNotFound
	case errors.Is(err, cherr.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, cherr.ErrUnauthorized), errors.Is(err, cherr.ErrSignRequired), errors.Is(err, cherr.ErrSignFailed):
		status = http.StatusUnauthorized
	case errors.Is(err, cherr.ErrAlreadyExists), errors.Is(err, cherr.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, cherr.ErrQuotaKeys), errors.Is(err, cherr.ErrQuotaBytes):
		status = http.StatusRequestEntityTooLarge
	case err == nil:
		status = http.StatusOK
	}
	msg := "error"
	if err != nil {
		msg = err.Error()
	}
	writeJSON(w, status, errBody{Error: msg})
}

func readJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return cherr.Wrap(cherr.ErrInvalidArg, err.Error())
	}
	return nil
}

func queryDuration(r *http.Request, name string, def time.Duration) time.Duration {
	s := r.URL.Query().Get(name)
	if s == "" {
		return def
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return def
	}
	if d < 0 {
		return def
	}
	if d > 60*time.Second {
		return 60 * time.Second
	}
	return d
}
