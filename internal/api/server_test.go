package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/LYH2263/go-confhub"
)

func TestHTTPPutGetWatch(t *testing.T) {
	h := confhub.New()
	t.Cleanup(func() { _ = h.Close() })
	s := New(h, Options{AllowCORS: true})
	ts := httptest.NewServer(s)
	t.Cleanup(ts.Close)

	body, _ := json.Marshal(map[string]any{"id": "prod", "owner": "alice"})
	res, err := http.Post(ts.URL+"/api/ns", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create ns %d", res.StatusCode)
	}
	_ = res.Body.Close()

	put, _ := json.Marshal(map[string]any{"payload": "hello-conf", "author": "alice"})
	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/api/ns/prod/keys/app.msg", bytes.NewReader(put))
	req.Header.Set("Content-Type", "application/json")
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("put %d %s", res.StatusCode, b)
	}
	_ = res.Body.Close()

	res, err = http.Get(ts.URL + "/api/ns/prod/keys/app.msg")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("get %d", res.StatusCode)
	}
}

func TestStaticIndex(t *testing.T) {
	dir := t.TempDir()
	web := filepath.Join(dir, "web")
	if err := os.MkdirAll(web, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(web, "index.html"), []byte("<html>ok</html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := confhub.New()
	t.Cleanup(func() { _ = h.Close() })
	s := New(h, Options{WebDir: web})
	ts := httptest.NewServer(s)
	t.Cleanup(ts.Close)
	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 || !bytes.Contains(b, []byte("ok")) {
		t.Fatalf("static %d %s", res.StatusCode, b)
	}
}
