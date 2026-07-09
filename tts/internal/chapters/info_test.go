package chapters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanFixture(t *testing.T) {
	dir := t.TempDir()
	// title block with vars
	write(t, dir, "0000_title_block.md", `---
title: "Test Book"
bundle: "the bundle"
---
`)
	write(t, dir, "0101_intro.md", `## Introduction

Hello {{.title}}.

This is {{.bundle}}.
`)
	// long-ish second chapter to force multiple chunks with tiny soft max
	body := strings.Repeat("Sentence number goes here. ", 200)
	write(t, dir, "0201_long.md", "## Long\n\n"+body)

	infos, vars, err := Scan(dir, 500)
	if err != nil {
		t.Fatal(err)
	}
	if vars["title"] != "Test Book" {
		t.Fatalf("vars: %v", vars)
	}
	if len(infos) != 3 {
		t.Fatalf("infos: %d", len(infos))
	}
	if !infos[0].Skip {
		t.Fatal("title block should skip")
	}
	if infos[1].Skip || infos[1].Chunks < 1 {
		t.Fatalf("intro: %+v", infos[1])
	}
	if !strings.Contains(FormatList(infos, 500), "0101_intro.md") {
		t.Fatal("format missing file")
	}
	// long chapter should split at soft 500
	if infos[2].Chunks < 2 {
		t.Fatalf("expected split for long chapter: %+v", infos[2])
	}
}

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
