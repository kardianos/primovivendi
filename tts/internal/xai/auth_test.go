package xai

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseKeyFile(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"xai-abc\n", "xai-abc"},
		{"# comment\nxai-xyz\n", "xai-xyz"},
		{"XAI_API_KEY=xai-eq\n", "xai-eq"},
		{"API_KEY=\"quoted\"\n", "quoted"},
		{"\n\n", ""},
	}
	for _, tc := range cases {
		if got := parseKeyFile([]byte(tc.in)); got != tc.want {
			t.Errorf("parseKeyFile(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestLoadAPIKeyFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ttskey")
	if err := os.WriteFile(path, []byte("xai-from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// clear env for this test
	t.Setenv("XAI_API_KEY", "")
	got, err := LoadAPIKey(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "xai-from-file" {
		t.Fatalf("got %q", got)
	}
}

func TestLoadAPIKeyEnvWins(t *testing.T) {
	t.Setenv("XAI_API_KEY", "xai-from-env")
	got, err := LoadAPIKey("/nonexistent/.ttskey")
	if err != nil {
		t.Fatal(err)
	}
	if got != "xai-from-env" {
		t.Fatalf("got %q", got)
	}
}
