package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type refWork struct {
	ID      int
	Year    int
	Country string
	Field   string
	Author  string
	Title   string
}

const refCap = 100000 // words kept per reference work

func loadManifest(root string) ([]refWork, error) {
	b, err := os.ReadFile(filepath.Join(root, "civ", "ngram", "reference.tsv"))
	if err != nil {
		return nil, err
	}
	var ws []refWork
	for _, line := range strings.Split(string(b), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		p := strings.Split(line, "\t")
		if len(p) != 6 {
			return nil, fmt.Errorf("reference.tsv: bad line %q", line)
		}
		id, err1 := strconv.Atoi(p[0])
		y, err2 := strconv.Atoi(p[1])
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("reference.tsv: bad numbers in %q", line)
		}
		ws = append(ws, refWork{id, y, p[2], p[3], p[4], p[5]})
	}
	return ws, nil
}

func refPath(data string, id int) string {
	return filepath.Join(data, "ref", fmt.Sprintf("pg%d.txt", id))
}

// cmdFetchRef downloads each work in the manifest that is not already saved,
// two seconds apart.
func cmdFetchRef(root, data string) error {
	ws, err := loadManifest(root)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(data, "ref"), 0o755); err != nil {
		return err
	}
	for _, w := range ws {
		out := refPath(data, w.ID)
		if _, err := os.Stat(out); err == nil {
			continue
		}
		u := fmt.Sprintf("https://www.gutenberg.org/cache/epub/%d/pg%d.txt", w.ID, w.ID)
		req, _ := http.NewRequest("GET", u, nil)
		req.Header.Set("User-Agent", userAgent)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		b, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("%s: %s", u, resp.Status)
		}
		if err := os.WriteFile(out, b, 0o644); err != nil {
			return err
		}
		log.Printf("pg%d %s: %d bytes", w.ID, w.Title, len(b))
		time.Sleep(2 * time.Second)
	}
	return nil
}

func loadRef(root, data string) ([]refWork, []*Doc, error) {
	ws, err := loadManifest(root)
	if err != nil {
		return nil, nil, err
	}
	docs := make([]*Doc, len(ws))
	for i, w := range ws {
		b, err := os.ReadFile(refPath(data, w.ID))
		if err != nil {
			return nil, nil, fmt.Errorf("%w (run: go run . fetch-ref)", err)
		}
		docs[i] = buildDoc(fmt.Sprintf("pg%d", w.ID), gutenbergParagraphs(string(b)), refCap)
	}
	return ws, docs, nil
}
