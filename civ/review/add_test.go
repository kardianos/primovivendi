package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddPacketPinsOffsetsAndRejectsStrictMiss(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "doc.md"), []byte("Alpha line.\n\nBeta line.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	packets := filepath.Join(root, "packets")

	raw := []byte(`{
  "version": 1,
  "items": [
    {"id": "call:doc:1", "path": "doc.md", "quote": "Beta line.", "problem": "test"}
  ]
}`)
	res, err := AddPacket(root, packets, raw, AddOptions{Slug: "ch01 calls", Pin: true, Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Items != 1 || len(res.Hits) != 1 || res.Hits[0].Status != StatusOffset || res.Hits[0].Offset != 13 {
		t.Fatalf("add result: %+v", res)
	}
	base := filepath.Base(res.PacketPath)
	if !strings.HasSuffix(base, "-ch01-calls.json") {
		t.Fatalf("unexpected filename: %s", base)
	}
	sess, err := LoadSession(root, packets, filepath.Join(root, "decisions.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sess.Items) != 1 || sess.Items[0].ByteOffset == nil || *sess.Items[0].ByteOffset != 13 {
		t.Fatalf("pinned offset not persisted: %+v", sess.Items)
	}

	// Strict rejects missing quotes.
	bad := []byte(`{"version":1,"items":[{"id":"miss","path":"doc.md","quote":"Nope."}]}`)
	if _, err := AddPacket(root, packets, bad, AddOptions{Slug: "bad", Strict: true}); err == nil {
		t.Fatal("expected strict mode to reject missing quote")
	}

	// Path traversal rejected.
	escape := []byte(`{"version":1,"items":[{"id":"esc","path":"../outside.md","quote":"x"}]}`)
	if _, err := AddPacket(root, packets, escape, AddOptions{Slug: "esc"}); err == nil {
		t.Fatal("expected path traversal rejection")
	}
}
