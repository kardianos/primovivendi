package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Concreteness ratings: M. Brysbaert, A. B. Warriner, and V. Kuperman,
// "Concreteness Ratings for 40 Thousand Generally Known English Word
// Lemmas," Behavior Research Methods 46 (2014), supplementary material 1.
// Each lemma is rated from 1 (abstract: known through words, such as
// "justice") to 5 (concrete: known through the senses, such as "hammer").
// The spreadsheet is downloaded to ext-ref/concreteness/ and its single
// words are written to a TSV beside it (neither is committed).

const concretenessURL = "https://static-content.springer.com/esm/art%3A10.3758%2Fs13428-013-0403-5/MediaObjects/13428_2013_403_MOESM1_ESM.xlsx"

func concretenessDir(data string) string { return filepath.Join(data, "concreteness") }

// cmdFetchConcreteness downloads the ratings, unless already saved, and
// writes word and mean rating to brysbaert_2014.tsv.
func cmdFetchConcreteness(data string) error {
	dir := concretenessDir(data)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	xlsx := filepath.Join(dir, "brysbaert_2014.xlsx")
	if _, err := os.Stat(xlsx); err != nil {
		req, _ := http.NewRequest("GET", concretenessURL, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; linear-review)")
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
			return fmt.Errorf("%s: %s", concretenessURL, resp.Status)
		}
		if err := os.WriteFile(xlsx, b, 0o644); err != nil {
			return err
		}
	}
	rows, err := readSheet(xlsx)
	if err != nil {
		return err
	}
	if len(rows) == 0 || len(rows[0]) < 3 || rows[0][0] != "Word" || rows[0][1] != "Bigram" || rows[0][2] != "Conc.M" {
		return fmt.Errorf("%s: unexpected columns %v", xlsx, rows[:min(1, len(rows))])
	}
	var sb strings.Builder
	n := 0
	for _, r := range rows[1:] {
		if len(r) < 3 || r[1] != "0" {
			continue // a two-word expression
		}
		x, err := strconv.ParseFloat(r[2], 64)
		if err != nil {
			continue
		}
		fmt.Fprintf(&sb, "%s\t%.2f\n", strings.ToLower(r[0]), x)
		n++
	}
	out := filepath.Join(dir, "brysbaert_2014.tsv")
	if err := os.WriteFile(out, []byte(sb.String()), 0o644); err != nil {
		return err
	}
	log.Printf("wrote %s: %d words", out, n)
	return nil
}

// readSheet returns the cells of the first worksheet as text, by row, with
// columns A, B, C... in order.
func readSheet(path string) ([][]string, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	read := func(name string, v any) error {
		f, err := zr.Open(name)
		if err != nil {
			return err
		}
		defer f.Close()
		return xml.NewDecoder(f).Decode(v)
	}
	var sst struct {
		SI []struct {
			T string `xml:"t"`
			R []struct {
				T string `xml:"t"`
			} `xml:"r"`
		} `xml:"si"`
	}
	if err := read("xl/sharedStrings.xml", &sst); err != nil {
		return nil, err
	}
	strs := make([]string, len(sst.SI))
	for i, si := range sst.SI {
		strs[i] = si.T
		for _, r := range si.R {
			strs[i] += r.T
		}
	}
	var sheet struct {
		Rows []struct {
			C []struct {
				R string `xml:"r,attr"`
				T string `xml:"t,attr"`
				V string `xml:"v"`
			} `xml:"c"`
		} `xml:"sheetData>row"`
	}
	if err := read("xl/worksheets/sheet1.xml", &sheet); err != nil {
		return nil, err
	}
	var rows [][]string
	for _, row := range sheet.Rows {
		var cells []string
		for _, c := range row.C {
			col := 0
			for _, ch := range c.R {
				if ch < 'A' || ch > 'Z' {
					break
				}
				col = col*26 + int(ch-'A'+1)
			}
			for len(cells) < col-1 {
				cells = append(cells, "")
			}
			v := c.V
			if c.T == "s" {
				i, err := strconv.Atoi(v)
				if err != nil || i < 0 || i >= len(strs) {
					return nil, fmt.Errorf("%s: bad shared string %q", path, v)
				}
				v = strs[i]
			}
			cells = append(cells, v)
		}
		rows = append(rows, cells)
	}
	return rows, nil
}

// loadConcreteness reads the ratings written by cmdFetchConcreteness.
func loadConcreteness(data string) (map[string]float64, error) {
	b, err := os.ReadFile(filepath.Join(concretenessDir(data), "brysbaert_2014.tsv"))
	if err != nil {
		return nil, fmt.Errorf("%w (run: go run . fetch-concreteness)", err)
	}
	m := map[string]float64{}
	for _, line := range strings.Split(string(b), "\n") {
		w, v, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		if x, err := strconv.ParseFloat(v, 64); err == nil {
			m[w] = x
		}
	}
	return m, nil
}
