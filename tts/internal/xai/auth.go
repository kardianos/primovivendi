package xai

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LoadAPIKey resolves the xAI API key from, in order:
//  1. XAI_API_KEY environment variable
//  2. A .ttskey file (first existing among candidates)
//
// The file may contain only the key, optionally with surrounding whitespace
// or a single trailing newline. Comments (# …) on their own lines are ignored.
func LoadAPIKey(extraPaths ...string) (string, error) {
	if k := strings.TrimSpace(os.Getenv("XAI_API_KEY")); k != "" {
		return k, nil
	}

	candidates := append([]string{}, extraPaths...)
	// Common locations relative to cwd (repo root or tts/)
	candidates = append(candidates,
		".ttskey",
		filepath.Join("..", ".ttskey"),
		filepath.Join("tts", ".ttskey"),
	)
	// Absolute: walk up from cwd looking for .ttskey
	if wd, err := os.Getwd(); err == nil {
		dir := wd
		for i := 0; i < 6; i++ {
			candidates = append(candidates, filepath.Join(dir, ".ttskey"))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	seen := map[string]bool{}
	for _, p := range candidates {
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if k := parseKeyFile(data); k != "" {
			return k, nil
		}
	}
	return "", fmt.Errorf("no API key: set XAI_API_KEY or create a .ttskey file")
}

func parseKeyFile(data []byte) string {
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// allow KEY=value form
		if i := strings.IndexByte(line, '='); i > 0 {
			name := strings.TrimSpace(line[:i])
			if strings.EqualFold(name, "XAI_API_KEY") || strings.EqualFold(name, "API_KEY") {
				return strings.TrimSpace(strings.Trim(line[i+1:], `"'`))
			}
		}
		return line
	}
	return ""
}

// NewClient builds a client from LoadAPIKey.
func NewClient(extraKeyPaths ...string) (*Client, error) {
	key, err := LoadAPIKey(extraKeyPaths...)
	if err != nil {
		return nil, err
	}
	return &Client{
		APIKey:     key,
		Endpoint:   DefaultEndpoint,
		HTTPClient: httpClient(),
	}, nil
}

func httpClient() *http.Client {
	return &http.Client{Timeout: 15 * time.Minute}
}
