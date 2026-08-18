package confhub

import (
	"testing"
)

func TestBug06_ExportSnapshotSharesLiveEntryBacking(t *testing.T) {
	h := New()
	t.Cleanup(func() { _ = h.Close() })
	if _, err := h.CreateNS("prod", "alice", 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Put("prod", "app.msg", []byte("hello-live"), PutOptions{Author: "alice"}); err != nil {
		t.Fatal(err)
	}
	blob, err := h.ExportSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	if blob == nil || len(blob.Entries) != 1 || len(blob.Entries[0].Versions) == 0 {
		t.Fatalf("export blob missing entry: %+v", blob)
	}
	blob.Entries[0].Versions[0].Payload[0] = 'J'
	got, err := h.Get("prod", "app.msg", ClientContext{})
	if err != nil {
		t.Fatal(err)
	}
	if string(got.Payload) != "hello-live" {
		t.Fatalf("exported snapshot must not share live payload; Get=%q", got.Payload)
	}

	h2 := New()
	t.Cleanup(func() { _ = h2.Close() })
	if err := h2.ImportSnapshot(blob); err != nil {
		t.Fatal(err)
	}
	blob.Entries[0].Versions[0].Payload[1] = 'X'
	got2, err := h2.Get("prod", "app.msg", ClientContext{})
	if err != nil {
		t.Fatal(err)
	}
	if got2.Payload[1] == 'X' {
		t.Fatalf("Import must copy buffers; live payload mutated to %q", got2.Payload)
	}
}
