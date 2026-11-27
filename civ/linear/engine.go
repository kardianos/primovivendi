package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"po/civ/sys/corenlp"
	"po/civ/sys/parser"
)

// The parser backend of civ/linear. "corenlp" reads the cached Stanford
// CoreNLP JSON under ext-ref/linear/out/ (and runs Java for new text);
// "go" parses in memory with the pure-Go pipeline of civ/sys/parser. Both
// produce the same corenlp.Doc structures with UTF-8 byte offsets, so the
// Rule 10 analysis below is shared.
const (
	engineCoreNLP = "corenlp"
	engineGo      = "go"
)

var goPipeline *parser.Pipeline

// pipelineGo loads the pure-Go parser once per process.
func pipelineGo(root string) (*parser.Pipeline, error) {
	if goPipeline != nil {
		return goPipeline, nil
	}
	p, err := parser.Load(root)
	if err != nil {
		return nil, fmt.Errorf("%w (train it with: civ/bin/parser train-all)", err)
	}
	goPipeline = p
	return p, nil
}

// validEngine reports whether engine names a parser backend.
func validEngine(engine string) bool { return engine == engineCoreNLP || engine == engineGo }

// loadDocEngine returns the parse of a source for the chosen engine.
func loadDocEngine(root, data string, s source, engine string) (*cnDoc, string, error) {
	if engine == engineGo {
		p, err := pipelineGo(root)
		if err != nil {
			return nil, "", err
		}
		return p.ParseDoc(s.Text), s.Text, nil
	}
	return loadParsed(data, s)
}

// parseTextEngine parses plain text for the chosen engine: in memory with
// the Go pipeline, or through the cached CoreNLP scratch directory.
func parseTextEngine(root, data, text, engine, mem string, threads int) (*cnDoc, string, error) {
	if engine == engineGo {
		p, err := pipelineGo(root)
		if err != nil {
			return nil, "", err
		}
		return p.ParseDoc(text), text, nil
	}
	return loadScratch(data, text, mem, threads)
}

// parseChapters selects the sources named in a comma-separated chapter list.
func parseChapters(list string) map[string]bool {
	out := map[string]bool{}
	for _, c := range strings.Split(list, ",") {
		if c = strings.TrimSpace(c); c != "" {
			out[c] = true
		}
	}
	return out
}

// cachedText finds the plain text CoreNLP parsed for name: a chapter or
// treatise under ext-ref/linear/in/, or a draft checked earlier under
// ext-ref/linear/scratch/ (name = hash). Both engines then read exactly the
// same text, whatever the state of the chapter files.
func cachedText(root, name string) (string, error) {
	for _, dir := range []string{"in", "scratch"} {
		b, err := os.ReadFile(filepath.Join(root, "ext-ref", "linear", dir, name+".txt"))
		if err == nil {
			return string(b), nil
		}
	}
	return "", fmt.Errorf("no cached CoreNLP input for %q", name)
}

// readCachedCoreNLP loads the cached CoreNLP JSON of name with UTF-8 byte offsets.
func readCachedCoreNLP(root, name string) (*cnDoc, error) {
	for _, dir := range []string{"out", "scratch"} {
		b, err := os.ReadFile(filepath.Join(root, "ext-ref", "linear", dir, name+".txt.json"))
		if err != nil {
			continue
		}
		var doc cnDoc
		if err := json.Unmarshal(b, &doc); err != nil {
			return nil, fmt.Errorf("corenlp json for %s: %w", name, err)
		}
		text, _ := cachedText(root, name)
		corenlp.NormalizeUTF16ToUTF8(&doc, text)
		return &doc, nil
	}
	return nil, fmt.Errorf("no cached CoreNLP parse for %q", name)
}

// sourceFromText makes a source whose paragraphs are found from the blank
// lines of the text (no Markdown line numbers).
func sourceFromText(name, text string) source {
	src := source{Name: name, Label: name, Text: text, Ref: true}
	src.ParaStart = []int{0}
	for i := 0; i+1 < len(src.Text); i++ {
		if src.Text[i] == '\n' && src.Text[i+1] == '\n' {
			src.ParaStart = append(src.ParaStart, i+2)
		}
	}
	return src
}

// cmdScores prints one JSON object per sentence: chapter, number, Markdown
// line, text, Rule 10 score, and findings. It feeds the parity report of
// `civ/bin/parser eval`, which runs it once per engine and aligns the two
// outputs by sentence text. Names may be chapters (ch13) or draft hashes.
func cmdScores(root, data, chapters, engine string) error {
	rates, err := loadConcreteness(data)
	if err != nil {
		return err
	}
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	enc := json.NewEncoder(out)
	names := []string{}
	for _, c := range strings.Split(chapters, ",") {
		if c = strings.TrimSpace(c); c != "" {
			names = append(names, c)
		}
	}
	if len(names) == 0 {
		return fmt.Errorf("scores: no chapters given (-chapter)")
	}
	var goP *parser.Pipeline
	if engine == engineGo {
		if goP, err = pipelineGo(root); err != nil {
			return err
		}
	}
	for _, name := range names {
		text, err := cachedText(root, name)
		if err != nil {
			return err
		}
		src := sourceFromText(name, text)
		var doc *cnDoc
		switch engine {
		case engineGo:
			doc = goP.ParseDoc(src.Text)
		default:
			if doc, err = readCachedCoreNLP(root, name); err != nil {
				return err
			}
		}
		for _, x := range analyzeDoc(&src, doc, src.Text, rates) {
			rec := struct {
				Chapter  string         `json:"chapter"`
				Num      int            `json:"num"`
				Line     int            `json:"line"`
				Text     string         `json:"text"`
				Score    float64        `json:"score"`
				Findings []scoreFinding `json:"findings"`
			}{name, x.Num, x.Line, x.Text, x.Score, nil}
			for _, f := range x.Findings {
				rec.Findings = append(rec.Findings, scoreFinding{Part: f.Part, Kind: f.Kind, Weight: f.Weight, Note: f.Note})
			}
			if err := enc.Encode(rec); err != nil {
				return err
			}
		}
	}
	return nil
}

type scoreFinding struct {
	Part   int     `json:"part"`
	Kind   string  `json:"kind"`
	Weight float64 `json:"weight"`
	Note   string  `json:"note"`
}
