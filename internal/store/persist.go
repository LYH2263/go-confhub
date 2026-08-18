package store

import (
	"encoding/json"
	"os"
	"path/filepath"

	cherr "github.com/LYH2263/go-confhub/internal/errors"
	"github.com/LYH2263/go-confhub/internal/meta"
)

type diskBlob struct {
	Entries []*meta.Entry `json:"entries"`
}

// AtomicWriteFile 先写临时文件再 rename，避免半截快照。
func AtomicWriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".confhub-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	ok := false
	defer func() {
		if !ok {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	ok = true
	return nil
}

func (m *Memory) SaveJSON(path string) error {
	entries := m.AllEntries()
	raw, err := json.MarshalIndent(diskBlob{Entries: entries}, "", "  ")
	if err != nil {
		return err
	}
	return AtomicWriteFile(path, raw)
}

func (m *Memory) LoadJSON(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cherr.Wrap(cherr.ErrNotFound, path)
		}
		return err
	}
	var blob diskBlob
	if err := json.Unmarshal(raw, &blob); err != nil {
		return cherr.Wrap(cherr.ErrSnapshot, err.Error())
	}
	return m.ReplaceAll(blob.Entries)
}

func EncodeEntriesJSON(entries []*meta.Entry) ([]byte, error) {
	return json.Marshal(diskBlob{Entries: entries})
}

func DecodeEntriesJSON(raw []byte) ([]*meta.Entry, error) {
	var blob diskBlob
	if err := json.Unmarshal(raw, &blob); err != nil {
		return nil, cherr.Wrap(cherr.ErrSnapshot, err.Error())
	}
	return blob.Entries, nil
}
