package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Stanford CoreNLP 4.6.0 with the English models jar, unpacked in
// ext-ref/corenlp/stanford-corenlp-4.6.0. The pipeline uses the shift-reduce
// constituency parser; its dependencies are converted from the tree. (In
// 4.6.0, running depparse together with parse gives some wrong token indices
// in the JSON, and neural coreference needs the constituency parse.)
var corenlpArgs = []string{
	"edu.stanford.nlp.pipeline.StanfordCoreNLP",
	"-annotators", "tokenize,ssplit,pos,lemma,ner,parse,coref",
	"-parse.model", "edu/stanford/nlp/models/srparser/englishSR.ser.gz",
	"-parse.maxlen", "150",
	"-coref.algorithm", "neural",
	"-ner.useSUTime", "false",
	"-ner.applyFineGrained", "false",
	"-ssplit.newlineIsSentenceBreak", "two",
	"-outputFormat", "json",
}

type cnDoc struct {
	Sentences []cnSent               `json:"sentences"`
	Corefs    map[string][]cnMention `json:"corefs"`
}

type cnSent struct {
	Index int     `json:"index"`
	Deps  []cnDep `json:"basicDependencies"`
	Toks  []cnTok `json:"tokens"`
}

type cnDep struct {
	Dep       string `json:"dep"`
	Governor  int    `json:"governor"`
	Dependent int    `json:"dependent"`
	DepGloss  string `json:"dependentGloss"`
}

type cnTok struct {
	Index int    `json:"index"`
	Word  string `json:"word"`
	Orig  string `json:"originalText"`
	Lemma string `json:"lemma"`
	Begin int    `json:"characterOffsetBegin"`
	End   int    `json:"characterOffsetEnd"`
	POS   string `json:"pos"`
	NER   string `json:"ner"`
}

type cnMention struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	Type      string `json:"type"`
	Start     int    `json:"startIndex"`
	End       int    `json:"endIndex"`
	Head      int    `json:"headIndex"`
	SentNum   int    `json:"sentNum"`
	IsRep     bool   `json:"isRepresentativeMention"`
	MentionID int    `json:"-"`
}

func workDir(data string) string { return filepath.Join(data, "linear") }

// cmdParse writes each source as plain text and runs CoreNLP on every source
// whose text changed since its last parse.
func cmdParse(root, data string, threads int, mem string) error {
	srcs, err := loadSources(root, data)
	if err != nil {
		return err
	}
	in := filepath.Join(workDir(data), "in")
	out := filepath.Join(workDir(data), "out")
	for _, d := range []string{in, out} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	var todo []string
	for _, s := range srcs {
		p := filepath.Join(in, s.Name+".txt")
		old, _ := os.ReadFile(p)
		_, jerr := os.Stat(filepath.Join(out, s.Name+".txt.json"))
		if bytes.Equal(old, []byte(s.Text)) && jerr == nil {
			continue
		}
		if err := os.WriteFile(p, []byte(s.Text), 0o644); err != nil {
			return err
		}
		todo = append(todo, p)
	}
	if len(todo) == 0 {
		log.Printf("all %d sources already parsed", len(srcs))
		return nil
	}
	list := filepath.Join(workDir(data), "filelist.txt")
	if err := os.WriteFile(list, []byte(strings.Join(todo, "\n")+"\n"), 0o644); err != nil {
		return err
	}
	cp := filepath.Join(data, "corenlp", "stanford-corenlp-4.6.0", "*")
	args := append([]string{"-mx" + mem, "-cp", cp}, corenlpArgs...)
	args = append(args, "-filelist", list, "-outputDirectory", out, "-threads", fmt.Sprint(threads))
	log.Printf("parsing %d sources with CoreNLP", len(todo))
	cmd := exec.Command("java", args...)
	logf, err := os.Create(filepath.Join(workDir(data), "corenlp.log"))
	if err != nil {
		return err
	}
	defer logf.Close()
	cmd.Stdout, cmd.Stderr = logf, logf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("corenlp: %w (see %s)", err, logf.Name())
	}
	return nil
}

func loadParsed(data string, s source) (*cnDoc, string, error) {
	b, err := os.ReadFile(filepath.Join(workDir(data), "out", s.Name+".txt.json"))
	if err != nil {
		return nil, "", fmt.Errorf("%w (run: go run . parse)", err)
	}
	var d cnDoc
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, "", err
	}
	text, err := os.ReadFile(filepath.Join(workDir(data), "in", s.Name+".txt"))
	if err != nil {
		return nil, "", err
	}
	if string(text) != s.Text {
		return nil, "", fmt.Errorf("%s changed since it was parsed (run: go run . parse)", s.Name)
	}
	return &d, string(text), nil
}
