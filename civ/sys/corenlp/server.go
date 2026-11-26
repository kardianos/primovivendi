package corenlp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"po/civ/sys/prof"
)

// Server manages an optional local StanfordCoreNLPServer HTTP daemon so that
// manually rewritten sentences can be re-parsed in ~30-50ms without JVM startup.
type Server struct {
	Root   string
	Port   int
	Mem    string
	Client *http.Client

	mu          sync.Mutex
	cmd         *exec.Cmd
	running     bool
	aliveCached bool
	aliveAt     time.Time
}

// NewServer creates a lazy-starting local CoreNLP server handle.
func NewServer(root string) *Server {
	return &Server{
		Root: root,
		Port: 9199,
		Mem:  "4g",
		Client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (s *Server) baseURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", s.Port)
}

// Alive reports whether a StanfordCoreNLPServer is currently responding
// (cached for 2 seconds to avoid repeated HTTP health checks during batch layout).
func (s *Server) Alive() bool {
	s.mu.Lock()
	if !s.aliveAt.IsZero() && time.Since(s.aliveAt) < 2*time.Second {
		v := s.aliveCached
		s.mu.Unlock()
		return v
	}
	s.mu.Unlock()
	ok := s.probeLive()
	s.mu.Lock()
	s.aliveCached = ok
	s.aliveAt = time.Now()
	s.mu.Unlock()
	return ok
}

func (s *Server) probeLive() bool {
	defer prof.Span("corenlp.probe_live")()
	c := &http.Client{Timeout: 400 * time.Millisecond}
	resp, err := c.Get(s.baseURL() + "/live")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// Start launches StanfordCoreNLPServer in the background if not already running.
func (s *Server) Start(ctx context.Context) error {
	defer prof.Span("corenlp.server_start")()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.probeLive() {
		s.running = true
		s.aliveCached = true
		s.aliveAt = time.Now()
		return nil
	}
	corenlpDir := filepath.Join(s.Root, "ext-ref", "corenlp", "stanford-corenlp-4.6.0")
	if _, err := os.Stat(corenlpDir); err != nil {
		return fmt.Errorf("corenlp not unpacked at %s: %w", corenlpDir, err)
	}
	if _, err := exec.LookPath("java"); err != nil {
		return fmt.Errorf("java not found: %w", err)
	}
	if s.Port == 0 {
		if ln, err := net.Listen("tcp", "127.0.0.1:0"); err == nil {
			s.Port = ln.Addr().(*net.TCPAddr).Port
			ln.Close()
		} else {
			s.Port = 9199
		}
	}
	cp := filepath.Join(corenlpDir, "*")
	// Avoid preloading the heavy PCFG parser on startup; CoreNLP loads the
	// shift-reduce model on the first HTTP request cleanly and binds the port fast.
	args := []string{
		"-mx" + s.Mem,
		"-cp", cp,
		"edu.stanford.nlp.pipeline.StanfordCoreNLPServer",
		"-port", fmt.Sprint(s.Port),
		"-timeout", "15000",
		"-threads", "2",
		"-quiet",
	}
	cmd := exec.Command("java", args...)
	logPath := filepath.Join(s.Root, "ext-ref", "linear", "corenlp_server.log")
	_ = os.MkdirAll(filepath.Dir(logPath), 0o755)
	if f, err := os.Create(logPath); err == nil {
		cmd.Stdout = f
		cmd.Stderr = f
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start corenlp server: %w", err)
	}
	s.cmd = cmd

	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		if s.probeLive() {
			s.running = true
			s.aliveCached = true
			s.aliveAt = time.Now()
			return nil
		}
		select {
		case <-ctx.Done():
			_ = s.Stop()
			return ctx.Err()
		case <-time.After(350 * time.Millisecond):
		}
	}
	return fmt.Errorf("corenlp server did not become ready within timeout (see %s)", logPath)
}

// Stop shuts down the managed server process, if this instance started it.
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	if s.cmd != nil && s.cmd.Process != nil {
		err := s.cmd.Process.Kill()
		_, _ = s.cmd.Process.Wait()
		s.cmd = nil
		return err
	}
	return nil
}

// ParseSentence parses a sentence via StanfordCoreNLPServer if reachable;
// otherwise it returns a deterministic heuristic dependency Tree so interactive
// workflows never block or fail when Java is offline.
func (s *Server) ParseSentence(ctx context.Context, text string) (*Tree, bool) {
	if s != nil && s.Alive() {
		if doc, err := s.query(ctx, text); err == nil {
			trees := BuildTrees(doc, text)
			if len(trees) > 0 {
				return trees[0], true
			}
		}
	}
	return ParseFallback(text), false
}

func (s *Server) query(ctx context.Context, text string) (*Doc, error) {
	detail := text
	if len(detail) > 48 {
		detail = detail[:48] + "..."
	}
	defer prof.SpanDetail("corenlp.http_query", detail)()
	props := `{"annotators":"tokenize,ssplit,pos,lemma,parse","parse.model":"edu/stanford/nlp/models/srparser/englishSR.ser.gz","outputFormat":"json"}`
	u := s.baseURL() + "/?properties=" + url.QueryEscape(props)
	req, err := http.NewRequestWithContext(ctx, "POST", u, strings.NewReader(text))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("corenlp HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var doc Doc
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}
	NormalizeUTF16ToUTF8(&doc, text)
	return &doc, nil
}

// ParseFallback builds a fast pure-Go heuristic POS and dependency Tree for a
// single sentence when CoreNLP is not running (e.g., offline or unit tests).
func ParseFallback(text string) *Tree {
	rawToks := fallbackTokenize(text)
	n := len(rawToks)
	if n == 0 {
		return &Tree{Text: strings.TrimSpace(text)}
	}
	t := &Tree{
		Toks: make([]Tok, n+1),
		Head: make([]int, n+1),
		Rel:  make([]string, n+1),
		Kids: make([][]int, n+1),
		Lo:   make([]int, n+1),
		Hi:   make([]int, n+1),
		Text: strings.TrimSpace(text),
	}
	for i, tk := range rawToks {
		tk.Index = i + 1
		tk.POS, tk.Lemma = guessPOSLemma(tk.Word, i == 0)
		t.Toks[i+1] = tk
	}
	// Choose root: first finite-looking main verb not inside an opening subordinate clause.
	root := 0
	for i := 1; i <= n; i++ {
		if strings.HasPrefix(t.Toks[i].POS, "VB") && !isAuxWord(strings.ToLower(t.Toks[i].Word)) {
			root = i
			break
		}
	}
	if root == 0 {
		for i := 1; i <= n; i++ {
			if strings.HasPrefix(t.Toks[i].POS, "VB") {
				root = i
				break
			}
		}
	}
	if root == 0 {
		root = 1
	}
	t.Root = root
	t.Head[root] = 0
	t.Rel[root] = "ROOT"

	// Attach tokens with clean heuristic dependencies:
	// - determiners/adjectives attach to the next noun
	// - auxiliaries attach to root
	// - prepositions attach to next noun (case) and that noun attaches to root (obl)
	// - first noun/pronoun before root becomes nsubj; first after root becomes obj
	seenSubj := false
	seenObj := false
	for i := 1; i <= n; i++ {
		if i == root {
			continue
		}
		pos := t.Toks[i].POS
		lw := strings.ToLower(t.Toks[i].Word)
		switch {
		case punctPOS[pos]:
			t.Head[i], t.Rel[i] = root, "punct"
		case pos == "DT" || pos == "PRP$":
			if nn := nextNoun(t, i); nn > 0 {
				t.Head[i], t.Rel[i] = nn, "det"
			} else {
				t.Head[i], t.Rel[i] = root, "det"
			}
		case strings.HasPrefix(pos, "JJ"):
			if nn := nextNoun(t, i); nn > 0 && nn-i <= 3 {
				t.Head[i], t.Rel[i] = nn, "amod"
			} else {
				t.Head[i], t.Rel[i] = root, "acomp"
			}
		case pos == "IN" || pos == "TO":
			if nn := nextNoun(t, i); nn > 0 && nn-i <= 4 {
				t.Head[i], t.Rel[i] = nn, "case"
			} else {
				t.Head[i], t.Rel[i] = root, "mark"
			}
		case pos == "MD" || isAuxWord(lw):
			t.Head[i], t.Rel[i] = root, "aux"
		case strings.HasPrefix(pos, "RB"):
			t.Head[i], t.Rel[i] = root, "advmod"
		case pos == "CC":
			t.Head[i], t.Rel[i] = root, "cc"
		case strings.HasPrefix(pos, "NN") || pos == "PRP":
			if i < root && !seenSubj && !precededByPrep(t, i) {
				t.Head[i], t.Rel[i] = root, "nsubj"
				seenSubj = true
			} else if i > root && !seenObj && !precededByPrep(t, i) {
				t.Head[i], t.Rel[i] = root, "obj"
				seenObj = true
			} else {
				t.Head[i], t.Rel[i] = root, "obl"
			}
		default:
			t.Head[i], t.Rel[i] = root, "dep"
		}
	}
	for i := 1; i <= n; i++ {
		if h := t.Head[i]; h > 0 {
			t.Kids[h] = append(t.Kids[h], i)
		}
	}
	var span func(i int) (int, int)
	visited := make([]bool, n+1)
	span = func(i int) (int, int) {
		if visited[i] {
			return i, i
		}
		visited[i] = true
		a, b := i, i
		for _, k := range t.Kids[i] {
			x, y := span(k)
			if x < a {
				a = x
			}
			if y > b {
				b = y
			}
		}
		t.Lo[i], t.Hi[i] = a, b
		return a, b
	}
	span(root)
	for i := 1; i <= n; i++ {
		if t.Lo[i] == 0 {
			t.Lo[i], t.Hi[i] = i, i
		}
	}
	t.Begin = t.Toks[1].Begin
	t.End = t.Toks[n].End
	return t
}

func nextNoun(t *Tree, from int) int {
	for j := from + 1; j <= t.N() && j <= from+4; j++ {
		p := t.Toks[j].POS
		if strings.HasPrefix(p, "NN") || p == "PRP" {
			return j
		}
		if punctPOS[p] || strings.HasPrefix(p, "VB") {
			return 0
		}
	}
	return 0
}

func precededByPrep(t *Tree, idx int) bool {
	for j := idx - 1; j >= 1 && j >= idx-3; j-- {
		p := t.Toks[j].POS
		if p == "IN" || p == "TO" {
			return true
		}
		if punctPOS[p] || strings.HasPrefix(p, "VB") {
			return false
		}
	}
	return false
}

func fallbackTokenize(text string) []Tok {
	var out []Tok
	n := len(text)
	for i := 0; i < n; {
		r, sz := utf8.DecodeRuneInString(text[i:])
		if unicode.IsSpace(r) {
			i += sz
			continue
		}
		start := i
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			i += sz
			for i < n {
				r2, sz2 := utf8.DecodeRuneInString(text[i:])
				if !(unicode.IsLetter(r2) || unicode.IsDigit(r2) || r2 == '-' || r2 == '\'') {
					break
				}
				i += sz2
			}
		} else {
			i += sz
		}
		w := text[start:i]
		out = append(out, Tok{Word: w, Orig: w, Begin: start, End: i, NER: "O"})
	}
	return out
}

var (
	fbDets  = map[string]bool{"the": true, "a": true, "an": true, "this": true, "that": true, "these": true, "those": true, "each": true, "every": true, "some": true, "any": true, "no": true, "such": true, "all": true, "both": true}
	fbPreps = map[string]bool{"in": true, "of": true, "for": true, "with": true, "on": true, "at": true, "from": true, "by": true, "about": true, "as": true, "into": true, "through": true, "after": true, "before": true, "between": true, "under": true, "over": true, "without": true, "within": true, "upon": true, "against": true, "because": true, "if": true, "when": true, "while": true, "though": true, "although": true, "unless": true, "until": true, "since": true}
	fbCC    = map[string]bool{"and": true, "or": true, "but": true, "nor": true, "yet": true, "so": true}
	fbPRP   = map[string]bool{"i": true, "we": true, "you": true, "he": true, "she": true, "it": true, "they": true, "me": true, "us": true, "him": true, "her": true, "them": true}
	fbPRPS  = map[string]bool{"my": true, "our": true, "your": true, "his": true, "its": true, "their": true}
	fbMD    = map[string]bool{"can": true, "could": true, "will": true, "would": true, "shall": true, "should": true, "may": true, "might": true, "must": true}
	fbAux   = map[string]bool{"is": true, "are": true, "was": true, "were": true, "be": true, "been": true, "being": true, "have": true, "has": true, "had": true, "do": true, "does": true, "did": true}
	fbVerbs = map[string]bool{"build": true, "builds": true, "built": true, "make": true, "makes": true, "made": true, "stand": true, "stands": true, "stood": true, "hold": true, "holds": true, "held": true, "keep": true, "keeps": true, "kept": true, "come": true, "comes": true, "came": true, "give": true, "gives": true, "gave": true, "take": true, "takes": true, "took": true, "fall": true, "falls": true, "fell": true, "fail": true, "fails": true, "failed": true, "grow": true, "grows": true, "grew": true, "know": true, "knows": true, "knew": true, "see": true, "sees": true, "saw": true, "live": true, "lives": true, "lived": true, "work": true, "works": true, "worked": true, "require": true, "requires": true, "depend": true, "depends": true, "carry": true, "carries": true, "break": true, "breaks": true, "broke": true}
	fbAdj   = map[string]bool{"good": true, "bad": true, "great": true, "small": true, "long": true, "short": true, "old": true, "new": true, "true": true, "false": true, "clear": true, "plain": true, "strong": true, "weak": true, "hard": true, "soft": true, "high": true, "low": true, "naked": true, "free": true, "just": true, "public": true, "private": true, "common": true, "human": true, "civil": true, "moral": true, "real": true, "whole": true, "open": true, "narrow": true, "broad": true}
)

func isAuxWord(lw string) bool { return fbAux[lw] || fbMD[lw] }

func guessPOSLemma(w string, first bool) (string, string) {
	switch w {
	case ".", "!", "?":
		return ".", w
	case ",":
		return ",", w
	case ";", ":", "—", "-":
		return ":", w
	case "(":
		return "-LRB-", w
	case ")":
		return "-RRB-", w
	case `"`, "``", "''":
		return "''", w
	}
	lw := strings.ToLower(w)
	switch {
	case fbDets[lw]:
		return "DT", lw
	case lw == "to":
		return "TO", lw
	case fbPreps[lw]:
		return "IN", lw
	case fbCC[lw]:
		return "CC", lw
	case fbPRP[lw]:
		return "PRP", lw
	case fbPRPS[lw]:
		return "PRP$", lw
	case fbMD[lw]:
		return "MD", lw
	case fbAux[lw]:
		if lw == "is" || lw == "has" || lw == "does" {
			return "VBZ", auxLemma(lw)
		}
		if lw == "was" || lw == "were" || lw == "had" || lw == "did" {
			return "VBD", auxLemma(lw)
		}
		return "VBP", auxLemma(lw)
	case fbVerbs[lw]:
		if strings.HasSuffix(lw, "s") && !strings.HasSuffix(lw, "ss") {
			return "VBZ", strings.TrimSuffix(lw, "s")
		}
		if strings.HasSuffix(lw, "ed") {
			return "VBD", strings.TrimSuffix(lw, "ed")
		}
		return "VBP", lw
	case fbAdj[lw]:
		return "JJ", lw
	case lw == "not" || lw == "never" || lw == "always" || lw == "also" || lw == "still" || (len(lw) > 4 && strings.HasSuffix(lw, "ly")):
		return "RB", lw
	case len(lw) > 4 && strings.HasSuffix(lw, "ing"):
		return "VBG", strings.TrimSuffix(lw, "ing")
	case len(lw) > 4 && strings.HasSuffix(lw, "ed"):
		return "VBD", strings.TrimSuffix(lw, "ed")
	case len(lw) > 4 && (strings.HasSuffix(lw, "ous") || strings.HasSuffix(lw, "ful") || strings.HasSuffix(lw, "able") || strings.HasSuffix(lw, "ible") || strings.HasSuffix(lw, "ive") || strings.HasSuffix(lw, "al")):
		return "JJ", lw
	case !first && len(w) > 1 && unicode.IsUpper(rune(w[0])):
		return "NNP", lw
	case len(lw) > 3 && strings.HasSuffix(lw, "s") && !strings.HasSuffix(lw, "ss") && !strings.HasSuffix(lw, "us") && !strings.HasSuffix(lw, "is"):
		return "NNS", strings.TrimSuffix(lw, "s")
	default:
		return "NN", lw
	}
}

func auxLemma(lw string) string {
	switch lw {
	case "is", "are", "was", "were", "be", "been", "being":
		return "be"
	case "have", "has", "had":
		return "have"
	case "do", "does", "did":
		return "do"
	default:
		return lw
	}
}
