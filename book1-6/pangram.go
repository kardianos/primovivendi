// Pangram AI-tone check for book1-6 chapters (publish text only).
//
// Usage:
//
//	go run pangram.go chapters/1100_axioms.md
//	go run pangram.go chapters/1000_questions.md chapters/1100_axioms.md
//	go run pangram.go -all chapters/
//	go run pangram.go -json build/pangram/out.json -all chapters/
//
// Auth (first match wins):
//
//	1. -key flag
//	2. PANGRAM_API_KEY env
//	3. .pangram_key in cwd or parents (one line, key only; gitignored)
//
// Prefer bulk API for multiple files (one job, shared poll). Soft style signal only.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	taskURL  = "https://text.external-api.pangram.com/task"
	bulkURL  = "https://text.external-api.pangram.com/bulk"
	pollEvery = 2 * time.Second
	pollMax   = 3 * time.Minute
)

var reHTMLComment = regexp.MustCompile(`(?s)<!--.*?-->`)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	keyFlag := flag.String("key", "", "API key (prefer .pangram_key or PANGRAM_API_KEY)")
	allDir := flag.String("all", "", "scan all *.md under this directory (non-recursive)")
	jsonOut := flag.String("json", "", "write full JSON results to this path")
	dashboard := flag.Bool("dashboard", false, "request public dashboard links")
	warnHuman := flag.Float64("min-human", 0.85, "warn if fraction_human is below this (0 disables)")
	flag.Parse()

	paths := flag.Args()
	if *allDir != "" {
		ents, err := os.ReadDir(*allDir)
		if err != nil {
			return err
		}
		for _, e := range ents {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			paths = append(paths, filepath.Join(*allDir, e.Name()))
		}
		sort.Strings(paths)
	}
	if len(paths) == 0 {
		return fmt.Errorf("usage: go run pangram.go [-all chapters/] [file.md ...]")
	}

	key, err := loadKey(*keyFlag)
	if err != nil {
		return err
	}

	type item struct {
		ID   string
		Path string
		Text string
		Words int
	}
	var items []item
	for _, p := range paths {
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		pub := publishText(raw)
		w := wordCount(pub)
		if strings.TrimSpace(pub) == "" {
			fmt.Fprintf(os.Stderr, "skip empty publish text: %s\n", p)
			continue
		}
		items = append(items, item{
			ID:    filepath.Base(p),
			Path:  p,
			Text:  pub,
			Words: w,
		})
	}
	if len(items) == 0 {
		return fmt.Errorf("no non-empty chapters to scan")
	}

	client := &http.Client{Timeout: 60 * time.Second}
	var results []map[string]any

	if len(items) == 1 {
		res, err := predictOne(client, key, items[0].Text, *dashboard)
		if err != nil {
			return err
		}
		res["_path"] = items[0].Path
		res["_id"] = items[0].ID
		res["_words"] = items[0].Words
		results = append(results, res)
	} else {
		bulkItems := make([]map[string]string, len(items))
		byID := map[string]item{}
		for i, it := range items {
			bulkItems[i] = map[string]string{"id": it.ID, "text": it.Text}
			byID[it.ID] = it
		}
		body := map[string]any{"items": bulkItems}
		raw, err := postJSON(client, bulkURL, key, body)
		if err != nil {
			return err
		}
		var launch struct {
			BulkID string `json:"bulk_id"`
			Status string `json:"status"`
		}
		if err := json.Unmarshal(raw, &launch); err != nil {
			return fmt.Errorf("bulk launch parse: %w\n%s", err, truncate(string(raw), 400))
		}
		if launch.BulkID == "" {
			return fmt.Errorf("bulk launch missing bulk_id: %s", truncate(string(raw), 400))
		}
		fmt.Fprintf(os.Stderr, "bulk_id=%s status=%s items=%d\n", launch.BulkID, launch.Status, len(items))

		if err := waitBulk(client, key, launch.BulkID); err != nil {
			return err
		}
		page, err := getJSON(client, fmt.Sprintf("%s/%s/results?offset=0&limit=1000", bulkURL, launch.BulkID), key)
		if err != nil {
			return err
		}
		var br struct {
			Items []struct {
				ID     string         `json:"id"`
				Stage  string         `json:"stage"`
				Error  *string        `json:"error"`
				Result map[string]any `json:"result"`
			} `json:"items"`
			Failed []struct {
				ID    string  `json:"id"`
				Error *string `json:"error"`
			} `json:"failed_items"`
		}
		if err := json.Unmarshal(page, &br); err != nil {
			return fmt.Errorf("bulk results parse: %w", err)
		}
		for _, fi := range br.Failed {
			msg := ""
			if fi.Error != nil {
				msg = *fi.Error
			}
			fmt.Fprintf(os.Stderr, "failed: %s %s\n", fi.ID, msg)
		}
		for _, it := range br.Items {
			if it.Result == nil {
				fmt.Fprintf(os.Stderr, "no result: %s stage=%s\n", it.ID, it.Stage)
				continue
			}
			meta, ok := byID[it.ID]
			if !ok {
				meta = item{ID: it.ID, Path: it.ID}
			}
			it.Result["_path"] = meta.Path
			it.Result["_id"] = meta.ID
			it.Result["_words"] = meta.Words
			results = append(results, it.Result)
		}
	}

	// Stable print order by path
	sort.Slice(results, func(i, j int) bool {
		return fmt.Sprint(results[i]["_path"]) < fmt.Sprint(results[j]["_path"])
	})

	for _, res := range results {
		printReport(res, *warnHuman)
	}

	if *jsonOut != "" {
		if err := os.MkdirAll(filepath.Dir(*jsonOut), 0o755); err != nil && filepath.Dir(*jsonOut) != "." {
			return err
		}
		b, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(*jsonOut, b, 0o600); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "wrote %s\n", *jsonOut)
	}
	return nil
}

func loadKey(flagKey string) (string, error) {
	if strings.TrimSpace(flagKey) != "" {
		return strings.TrimSpace(flagKey), nil
	}
	if k := strings.TrimSpace(os.Getenv("PANGRAM_API_KEY")); k != "" {
		return k, nil
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		p := filepath.Join(dir, ".pangram_key")
		b, err := os.ReadFile(p)
		if err == nil {
			k := strings.TrimSpace(string(b))
			if k == "" {
				return "", fmt.Errorf("%s is empty", p)
			}
			// first line only
			if i := strings.IndexAny(k, "\r\n"); i >= 0 {
				k = strings.TrimSpace(k[:i])
			}
			return k, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("no API key: set PANGRAM_API_KEY or create .pangram_key (gitignored)")
}

// publishText mirrors concat.go new-mode line markers + HTML comment strip.
func publishText(b []byte) string {
	b, _ = stripComments(b)
	b, _ = applyLineMarkers(b)
	// collapse excess blank lines
	lines := strings.Split(string(b), "\n")
	var out []string
	blank := 0
	for _, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			blank++
			if blank <= 2 {
				out = append(out, "")
			}
			continue
		}
		blank = 0
		out = append(out, ln)
	}
	return strings.TrimSpace(strings.Join(out, "\n")) + "\n"
}

func stripComments(b []byte) ([]byte, int) {
	n := len(reHTMLComment.FindAll(b, -1))
	if n == 0 {
		return b, 0
	}
	return reHTMLComment.ReplaceAll(b, nil), n
}

func applyLineMarkers(b []byte) ([]byte, struct{}) {
	var st struct{}
	if len(b) == 0 {
		return b, st
	}
	lines := bytes.Split(b, []byte("\n"))
	out := make([][]byte, 0, len(lines))
	for _, line := range lines {
		s := string(bytes.TrimSuffix(line, []byte("\r")))
		switch {
		case strings.HasPrefix(s, "%%"):
			continue
		case strings.HasPrefix(s, "%+"):
			body := s[len("%+"):]
			if strings.HasPrefix(body, " ") {
				body = body[1:]
			}
			out = append(out, []byte(body))
		case strings.HasPrefix(s, "%-"):
			continue
		default:
			out = append(out, []byte(s))
		}
	}
	return bytes.Join(out, []byte("\n")), st
}

func wordCount(s string) int {
	n := 0
	for _, f := range strings.Fields(s) {
		if f != "" {
			n++
		}
	}
	return n
}

func predictOne(client *http.Client, key, text string, dashboard bool) (map[string]any, error) {
	body := map[string]any{"text": text, "public_dashboard_link": dashboard}
	raw, err := postJSON(client, taskURL, key, body)
	if err != nil {
		return nil, err
	}
	var launch struct {
		TaskID string `json:"task_id"`
	}
	if err := json.Unmarshal(raw, &launch); err != nil || launch.TaskID == "" {
		return nil, fmt.Errorf("task launch: %w body=%s", err, truncate(string(raw), 300))
	}
	deadline := time.Now().Add(pollMax)
	for time.Now().Before(deadline) {
		raw, err := getJSON(client, taskURL+"/"+launch.TaskID, key)
		if err != nil {
			return nil, err
		}
		var stage struct {
			Stage string `json:"stage"`
		}
		_ = json.Unmarshal(raw, &stage)
		switch stage.Stage {
		case "STAGE_SUCCESS":
			var m map[string]any
			if err := json.Unmarshal(raw, &m); err != nil {
				return nil, err
			}
			return m, nil
		case "STAGE_FAILED":
			return nil, fmt.Errorf("task failed: %s", truncate(string(raw), 400))
		default:
			time.Sleep(pollEvery)
		}
	}
	return nil, fmt.Errorf("task poll timeout for %s", launch.TaskID)
}

func waitBulk(client *http.Client, key, bulkID string) error {
	deadline := time.Now().Add(pollMax)
	url := bulkURL + "/" + bulkID
	for time.Now().Before(deadline) {
		raw, err := getJSON(client, url, key)
		if err != nil {
			return err
		}
		var st struct {
			Status    string `json:"status"`
			Succeeded int    `json:"succeeded"`
			Failed    int    `json:"failed"`
			Total     int    `json:"total_items"`
		}
		if err := json.Unmarshal(raw, &st); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "bulk status=%s succeeded=%d failed=%d total=%d\n", st.Status, st.Succeeded, st.Failed, st.Total)
		switch st.Status {
		case "succeeded", "partial", "failed":
			if st.Status == "failed" && st.Succeeded == 0 {
				return fmt.Errorf("bulk failed: %s", truncate(string(raw), 400))
			}
			return nil
		default:
			time.Sleep(pollEvery)
		}
	}
	return fmt.Errorf("bulk poll timeout for %s", bulkID)
}

func postJSON(client *http.Client, url, key string, body any) ([]byte, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", key)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("POST %s: %s\n%s", url, resp.Status, truncate(string(raw), 500))
	}
	return raw, nil
}

func getJSON(client *http.Client, url, key string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", key)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GET %s: %s\n%s", url, resp.Status, truncate(string(raw), 500))
	}
	return raw, nil
}

func printReport(res map[string]any, minHuman float64) {
	path := fmt.Sprint(res["_path"])
	words := res["_words"]
	fmt.Printf("\n== %s  (publish words≈%v) ==\n", path, words)
	fmt.Printf("headline: %v\n", res["headline"])
	fmt.Printf("short:    %v\n", res["prediction_short"])
	if p, ok := res["prediction"].(string); ok && p != "" {
		fmt.Printf("detail:   %s\n", p)
	}
	fh, fa, faa := asFloat(res["fraction_human"]), asFloat(res["fraction_ai"]), asFloat(res["fraction_ai_assisted"])
	fmt.Printf("fractions: human=%.3f  ai_assisted=%.3f  ai=%.3f\n", fh, faa, fa)
	if minHuman > 0 && fh < minHuman {
		fmt.Printf("WARN: fraction_human %.3f < %.3f (soft style signal)\n", fh, minHuman)
	}
	short := fmt.Sprint(res["prediction_short"])
	if short != "" && !strings.EqualFold(short, "Human") {
		fmt.Printf("WARN: prediction_short=%s (soft style signal)\n", short)
	}
	if link, ok := res["dashboard_link"].(string); ok && link != "" {
		fmt.Printf("dashboard: %s\n", link)
	}

	windows, _ := res["windows"].([]any)
	if len(windows) == 0 {
		return
	}
	fmt.Println("flagged windows:")
	n := 0
	for _, w := range windows {
		wm, ok := w.(map[string]any)
		if !ok {
			continue
		}
		label := fmt.Sprint(wm["label"])
		score := asFloat(wm["ai_assistance_score"])
		if strings.EqualFold(label, "Human") && score < 0.35 {
			continue
		}
		text := fmt.Sprint(wm["text"])
		text = strings.Join(strings.Fields(text), " ")
		if len(text) > 160 {
			text = text[:157] + "..."
		}
		fmt.Printf("  [%v-%v] %s  conf=%v  score=%.2f\n    %q\n",
			wm["start_index"], wm["end_index"], label, wm["confidence"], score, text)
		n++
	}
	if n == 0 {
		fmt.Println("  (none above soft threshold)")
	}
}

func asFloat(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case json.Number:
		f, _ := t.Float64()
		return f
	case int:
		return float64(t)
	default:
		return 0
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
