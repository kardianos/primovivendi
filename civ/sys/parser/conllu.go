// conllu.go — corpus readers and label normalizers for training: the
// CoNLL-U reader for Universal Dependencies English-EWT, the CoreNLP JSON
// reader for the silver distillation data, and the canonical POS tag and
// dependency label inventories that civ/compose and civ/linear consume.
package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"po/civ/sys/corenlp"
)

// PTBTags is the canonical Penn Treebank XPOS inventory this parser emits, in
// a fixed order (class order of the POS tagger).
var PTBTags = []string{
	"CC", "CD", "DT", "EX", "FW", "IN", "JJ", "JJR", "JJS", "LS", "MD", "NN", "NNS", "NNP", "NNPS",
	"PDT", "POS", "PRP", "PRP$", "RB", "RBR", "RBS", "RP", "SYM", "TO", "UH", "VB", "VBD", "VBG",
	"VBN", "VBP", "VBZ", "WDT", "WP", "WP$", "WRB", "``", "''", ",", ".", ":", "-LRB-", "-RRB-",
	"HYPH", "$", "#",
}

// DepLabels is the canonical dependency label inventory (UD v2 English as
// emitted by CoreNLP 4.6.0), ROOT first.
var DepLabels = []string{
	"ROOT",
	"nsubj", "nsubj:pass", "csubj", "csubj:pass", "expl",
	"obj", "iobj", "obl", "obl:npmod", "obl:tmod", "ccomp", "xcomp",
	"advcl", "acl", "acl:relcl", "parataxis",
	"nmod", "nmod:poss", "amod", "advmod", "appos", "nummod", "compound", "compound:prt", "flat", "fixed",
	"det", "det:predet", "case", "mark", "aux", "aux:pass", "cop", "cc", "cc:preconj", "conj",
	"punct", "discourse", "list", "orphan", "vocative", "dep",
}

// NERLabels is the canonical NER label set written to Tok.NER.
var NERLabels = []string{"O", "PERSON", "ORG", "LOC", "MISC", "NUMBER", "DATE"}

var (
	ptbIndex    = map[string]int{}
	depIndex    = map[string]int{}
	nerIndex    = map[string]int{}
	depSubtypes = map[string]string{ // subtypes kept as-is; other subtypes fall back to the base relation
		"nsubj:pass": "nsubj:pass", "csubj:pass": "csubj:pass", "aux:pass": "aux:pass", "acl:relcl": "acl:relcl",
		"nmod:poss": "nmod:poss", "obl:npmod": "obl:npmod", "obl:tmod": "obl:tmod", "compound:prt": "compound:prt",
		"cc:preconj": "cc:preconj", "det:predet": "det:predet",
	}
)

func init() {
	for i, t := range PTBTags {
		ptbIndex[t] = i
	}
	for i, l := range DepLabels {
		depIndex[l] = i
	}
	for i, l := range NERLabels {
		nerIndex[l] = i
	}
}

// PTBIndex returns the class index of a canonical POS tag (-1 if unknown).
func PTBIndex(tag string) int { return ptbIndex[tag] }

// DepIndex returns the index of a canonical dependency label (-1 if unknown).
func DepIndex(label string) int { return depIndex[label] }

// IsCanonicalTag reports whether tag is in the canonical POS inventory.
func IsCanonicalTag(tag string) bool { _, ok := ptbIndex[tag]; return ok }

// NormDeprel maps a Universal Dependencies relation to the canonical label
// inventory: "root" becomes "ROOT", listed subtypes are kept, and every
// other subtype is reduced to its base relation (unknown relations become "dep").
func NormDeprel(rel string) string {
	if rel == "root" || rel == "ROOT" {
		return "ROOT"
	}
	if _, ok := depIndex[rel]; ok {
		return rel
	}
	if base, ok := depSubtypes[rel]; ok {
		return base
	}
	if i := strings.IndexByte(rel, ':'); i > 0 {
		base := rel[:i]
		if _, ok := depIndex[base]; ok {
			return base
		}
	}
	return "dep"
}

// NormXPOS maps a treebank XPOS to the canonical Penn Treebank inventory.
// upos is the CoNLL-U UPOS column, used for the few treebank-only tags
// (GW, NFP, ADD, AFX).
func NormXPOS(xpos, upos string) string {
	if _, ok := ptbIndex[xpos]; ok {
		return xpos
	}
	switch xpos {
	case "AFX":
		return "HYPH"
	case "ADD":
		return "NNP"
	case "NFP": // non-first punctuation: dashes, bullets, ellipses; emoticons stay SYM
		if upos == "SYM" {
			return "SYM"
		}
		return ":"
	case "GW": // a "goeswith" fragment takes the tag of its part of speech
		switch upos {
		case "NOUN":
			return "NN"
		case "PROPN":
			return "NNP"
		case "VERB", "AUX":
			return "VB"
		case "ADJ":
			return "JJ"
		case "ADV":
			return "RB"
		case "ADP", "SCONJ":
			return "IN"
		case "CCONJ":
			return "CC"
		case "DET":
			return "DT"
		case "PRON":
			return "PRP"
		case "PART":
			return "TO"
		case "NUM":
			return "CD"
		case "SYM":
			return "SYM"
		case "INTJ":
			return "UH"
		case "PUNCT":
			return "."
		}
		return "."
	}
	// Penn Treebank bracket aliases.
	switch xpos {
	case "-LSB-", "-LCB-":
		return "-LRB-"
	case "-RSB-", "-RCB-":
		return "-RRB-"
	}
	return "NN"
}

// NormNER maps a CoreNLP NER label to the canonical set.
func NormNER(label string) string {
	switch label {
	case "PERSON":
		return "PERSON"
	case "ORG", "ORGANIZATION":
		return "ORG"
	case "LOC", "LOCATION":
		return "LOC"
	case "MISC":
		return "MISC"
	case "NUMBER", "ORDINAL", "MONEY", "PERCENT", "QUANTITY":
		return "NUMBER"
	case "DATE", "TIME", "DURATION", "SET":
		return "DATE"
	}
	return "O"
}

// Sentence is a pre-tokenized sentence with optional gold annotations. All
// per-token slices have length n (0-based); Head/Rel are indexed by token
// position with 1-based governor numbers inside.
type Sentence struct {
	Words  []string // surface forms
	POS    []string // gold Penn Treebank tags (nil if unknown)
	Lemma  []string // optional
	NER    []string // optional canonical NER labels
	Head   []int    // gold governor of token i (0 = root), 1-based numbers; nil if unknown
	Rel    []string // gold canonical dependency labels
	Src    string   // corpus name (e.g. "ch01", "ewt-train")
	Weight int      // copies per epoch when mixing corpora (0 and 1 mean once)
}

func (s *Sentence) Len() int { return len(s.Words) }

// ---- CoNLL-U -----------------------------------------------------------------

// ReadCoNLLU parses a CoNLL-U file, skipping multi-word token ranges ("1-2")
// and empty nodes ("1.1"). Labels are normalized to the canonical inventory.
func ReadCoNLLU(path string) ([]*Sentence, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseCoNLLU(f, strings.TrimSuffix(filepath.Base(path), ".conllu"))
}

func parseCoNLLU(r io.Reader, src string) ([]*Sentence, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<16), 1<<20)
	var out []*Sentence
	var cur *Sentence
	flush := func() {
		if cur != nil && cur.Len() > 0 {
			out = append(out, cur)
		}
		cur = nil
	}
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			flush()
			continue
		}
		if line[0] == '#' {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 8 {
			continue
		}
		id := fields[0]
		if strings.ContainsAny(id, "-.") {
			continue
		}
		if cur == nil {
			cur = &Sentence{Src: src}
		}
		head, _ := strconv.Atoi(fields[6])
		cur.Words = append(cur.Words, fields[1])
		cur.Lemma = append(cur.Lemma, fields[2])
		cur.POS = append(cur.POS, NormXPOS(fields[4], fields[3]))
		cur.Head = append(cur.Head, head)
		cur.Rel = append(cur.Rel, NormDeprel(fields[7]))
	}
	flush()
	return out, sc.Err()
}

// ---- CoreNLP JSON (silver distillation data) ----------------------------------

// Sentence types from a CoreNLP JSON document; a nil *Sentence is appended for
// a malformed sentence (bad token indices or glosses), so positions match.
func sentencesFromCoreNLP(doc *corenlp.Doc, src string) ([]*Sentence, error) {
	out := make([]*Sentence, 0, len(doc.Sentences))
	for _, cs := range doc.Sentences {
		n := len(cs.Toks)
		if n == 0 {
			out = append(out, nil)
			continue
		}
		s := &Sentence{Src: src, Words: make([]string, n), POS: make([]string, n), Head: make([]int, n), Rel: make([]string, n)}
		good := true
		for i, tk := range cs.Toks {
			// CoreNLP's JSON keeps word == originalText; the surface form is used.
			s.Words[i] = tk.Word
			if tk.Orig != "" {
				s.Words[i] = tk.Orig
			}
			s.POS[i] = NormXPOS(tk.POS, "")
			if !IsCanonicalTag(s.POS[i]) {
				good = false
			}
			s.Head[i] = -1
		}
		if !good {
			out = append(out, nil)
			continue
		}
		roots := 0
		for _, d := range cs.Deps {
			if d.Dependent < 1 || d.Dependent > n || d.Governor < 0 || d.Governor > n {
				good = false
				break
			}
			if d.DepGloss != "" && cs.Toks[d.Dependent-1].Word != d.DepGloss {
				good = false
			}
			if s.Head[d.Dependent-1] != -1 {
				good = false
			}
			s.Head[d.Dependent-1] = d.Governor
			s.Rel[d.Dependent-1] = NormDeprel(d.Dep)
			if d.Governor == 0 {
				roots++
			}
		}
		if !good || roots != 1 {
			out = append(out, nil)
			continue
		}
		for _, h := range s.Head {
			if h == -1 {
				good = false
				break
			}
		}
		if !good {
			out = append(out, nil)
			continue
		}
		s.NER = make([]string, n)
		for i, tk := range cs.Toks {
			s.NER[i] = NormNER(tk.NER)
		}
		out = append(out, s)
	}
	return out, nil
}

// ReadCoreNLPDoc loads a cached CoreNLP JSON file (raw UTF-16 offsets are not
// needed for training, so no normalization is required here).
func ReadCoreNLPDoc(path, src string) ([]*Sentence, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc corenlp.Doc
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("corenlp json %s: %w", path, err)
	}
	return sentencesFromCoreNLP(&doc, src)
}

// HeldoutChapters lists the CoreNLP-silver chapters reserved for evaluation.
// Nothing in this package trains on them.
var HeldoutChapters = []string{"ch13", "ch14"}

// BookChapters lists the chapters usable as silver training data (ch01..ch12).
var BookChapters = []string{"ch01", "ch02", "ch03", "ch04", "ch05", "ch06", "ch07", "ch08", "ch09", "ch10", "ch11", "ch12"}

// TreatiseIDs lists the eight reference treatises that make up the silver
// distillation set.
var TreatiseIDs = []string{"pg4351", "pg4367", "pg6435", "pg935", "pg13529", "pg26334", "pg34901", "pg35916"}

// IsHeldout reports whether a corpus source name is reserved for evaluation.
func IsHeldout(name string) bool { return name == "ch13" || name == "ch14" }

// TrainingSources returns the silver CoreNLP sources allowed for training:
// the eight treatises and ch01..ch12. ch13 and ch14 are never returned.
func TrainingSources() []string {
	out := make([]string, 0, len(TreatiseIDs)+len(BookChapters))
	out = append(out, TreatiseIDs...)
	for _, c := range BookChapters {
		out = append(out, c)
	}
	return out
}

// SilverDir returns ext-ref/linear/out relative to root.
func SilverDir(root string) string { return filepath.Join(root, "ext-ref", "linear", "out") }

// LoadSilver loads the cached CoreNLP parses of the named sources from
// ext-ref/linear/out. Held-out chapters are refused.
func LoadSilver(root string, names []string) ([]*Sentence, error) {
	var out []*Sentence
	for _, name := range names {
		if IsHeldout(name) {
			return nil, fmt.Errorf("parser: %s is a held-out evaluation chapter and cannot be used for training", name)
		}
		path := filepath.Join(SilverDir(root), name+".txt.json")
		ss, err := ReadCoreNLPDoc(path, name)
		if err != nil {
			return nil, err
		}
		out = append(out, ss...)
	}
	return out, nil
}

// LoadEWT loads a UD_English-EWT split (train/dev/test) from
// ext-ref/parser/ud_ewt/.
func LoadEWT(root, split string) ([]*Sentence, error) {
	path := filepath.Join(root, "ext-ref", "parser", "ud_ewt", "en_ewt-ud-"+split+".conllu")
	ss, err := ReadCoNLLU(path)
	if err != nil {
		return nil, fmt.Errorf("parser: UD_English-EWT %s not found (run: civ/bin/parser fetch-ud): %w", split, err)
	}
	return ss, nil
}

// LoadGUM loads a UD_English-GUM split (train/dev/test) from ext-ref/parser/ud_gum/.
func LoadGUM(root, split string) ([]*Sentence, error) {
	path := filepath.Join(root, "ext-ref", "parser", "ud_gum", "en_gum-ud-"+split+".conllu")
	ss, err := ReadCoNLLU(path)
	if err != nil {
		return nil, fmt.Errorf("parser: UD_English-GUM %s not found (run: civ/bin/parser fetch-ud): %w", split, err)
	}
	return ss, nil
}

// EWTAvailable reports whether the UD_English-EWT treebank is downloaded.
func EWTAvailable(root string) bool {
	_, err := os.Stat(filepath.Join(root, "ext-ref", "parser", "ud_ewt", "en_ewt-ud-train.conllu"))
	return err == nil
}

// ---- Round 2: dev split, optional treebank, sealed drafts ---------------------

// TrainChapters lists the book chapters used as silver training data in
// Round 2: ch12 is held out from training as the in-domain dev set.
var TrainChapters = []string{"ch01", "ch02", "ch03", "ch04", "ch05", "ch06", "ch07", "ch08", "ch09", "ch10", "ch11"}

// DevBookChapter is the chapter held out from Round-2 training for
// checkpoint selection and the Stage 4 tuning (dev_book).
const DevBookChapter = "ch12"

// UDDir returns the directory of a UD treebank under ext-ref/parser/.
func UDDir(root, name string) string { return filepath.Join(root, "ext-ref", "parser", name) }

// LoadLinES loads a UD_English-LinES split (train, dev, test) from
// ext-ref/parser/ud_lines/.
func LoadLinES(root, split string) ([]*Sentence, error) {
	path := filepath.Join(UDDir(root, "ud_lines"), "en_lines-ud-"+split+".conllu")
	ss, err := ReadCoNLLU(path)
	if err != nil {
		return nil, fmt.Errorf("parser: UD_English-LinES %s not found (run: civ/bin/parser fetch-ud): %w", split, err)
	}
	for _, s := range ss {
		s.Src = "lines-" + s.Src
	}
	return ss, nil
}

// LinESAvailable reports whether the optional UD_English-LinES treebank is
// downloaded.
func LinESAvailable(root string) bool {
	_, err := os.Stat(filepath.Join(UDDir(root, "ud_lines"), "en_lines-ud-train.conllu"))
	return err == nil
}

// GUMAvailable reports whether UD_English-GUM is downloaded.
func GUMAvailable(root string) bool {
	_, err := os.Stat(filepath.Join(UDDir(root, "ud_gum"), "en_gum-ud-train.conllu"))
	return err == nil
}

// DraftDir returns the directory of the CoreNLP parses of drafts
// (ext-ref/linear/scratch), which holds name.txt and name.txt.json pairs.
func DraftDir(root string) string { return filepath.Join(root, "ext-ref", "linear", "scratch") }

// HoldoutSealKeys returns the sentence keys of the held-out chapters' raw
// text. It is the only function besides evaluation that reads ch13 or ch14,
// and it reads nothing but the text: the keys exist to exclude, never to
// train on. A draft that shares one sentence with a held-out chapter is
// sealed.
func HoldoutSealKeys(root string) (map[string]bool, error) {
	keys := map[string]bool{}
	for _, name := range HeldoutChapters {
		b, err := os.ReadFile(filepath.Join(root, "ext-ref", "linear", "in", name+".txt"))
		if err != nil {
			return nil, fmt.Errorf("parser: cannot read the held-out chapter %s for sealing drafts: %w", name, err)
		}
		for k := range sentenceKeys(string(b)) {
			keys[k] = true
		}
	}
	return keys, nil
}

// DevSealKeys returns the sentence keys of dev_book's raw text (ch12). It
// reads nothing but the text, as HoldoutSealKeys does, and for the same
// purpose: a draft that shares a sentence with the development chapter is
// excluded from training, so that dev_book measures what the held-out
// chapters will, whose overlapping drafts are sealed too.
func DevSealKeys(root string) (map[string]bool, error) {
	b, err := os.ReadFile(filepath.Join(root, "ext-ref", "linear", "in", DevBookChapter+".txt"))
	if err != nil {
		return nil, fmt.Errorf("parser: cannot read %s for sealing drafts: %w", DevBookChapter, err)
	}
	keys := map[string]bool{}
	for k := range sentenceKeys(string(b)) {
		keys[k] = true
	}
	return keys, nil
}

// SilverSentenceKeys returns the sentence keys (CoreNLP tokenization) of
// the named cached sources, for deduplicating drafts against them.
func SilverSentenceKeys(root string, names []string) (map[string]bool, error) {
	keys := map[string]bool{}
	for _, name := range names {
		ss, err := ReadCoreNLPDoc(filepath.Join(SilverDir(root), name+".txt.json"), name)
		if err != nil {
			return nil, err
		}
		for _, s := range ss {
			if s == nil {
				continue
			}
			if k := sentenceKeyToks(s.Words); utf8.RuneCountInString(k) >= minSharedRunes {
				keys[k] = true
			}
		}
	}
	return keys, nil
}

// sentenceKeyToks is sentenceKey over surface forms.
func sentenceKeyToks(words []string) string {
	return strings.ToLower(strings.Join(words, " "))
}

// DraftLoad reports what LoadSealedDrafts did.
type DraftLoad struct {
	Files, Unreadable, Loaded int
	Sealed                    int // drafts excluded for sharing a sentence with ch13 or ch14
	SealedDev                 int // drafts excluded for sharing a sentence with ch12
	Deduped, Sentences        int
}

// LoadSealedDrafts loads the CoreNLP parses of the drafts in dir as silver
// sentences. A draft whose text shares a sentence (minSharedRunes) with a
// key of seal is excluded whole, and so is a draft without a parse. dev
// (nil: none) holds the keys of the development chapter; a draft sharing a
// sentence with it is excluded too and counted apart. Within the remaining
// drafts, sentences that occur verbatim in reference (the chapters and
// treatises already in the corpus) are dropped, so every returned sentence
// is new in-domain text. The result is sorted by file name for a
// reproducible order.
func LoadSealedDrafts(dir string, seal, dev, reference map[string]bool) ([]*Sentence, *DraftLoad, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.txt.json"))
	if err != nil {
		return nil, nil, err
	}
	sort.Strings(paths)
	var out []*Sentence
	info := &DraftLoad{Files: len(paths)}
	for _, path := range paths {
		base := strings.TrimSuffix(filepath.Base(path), ".txt.json")
		text, err := os.ReadFile(filepath.Join(dir, base+".txt"))
		if err != nil {
			info.Unreadable++
			continue
		}
		keys := sentenceKeys(string(text))
		if overlapsSentence(keys, seal) {
			info.Sealed++
			continue
		}
		if dev != nil && overlapsSentence(keys, dev) {
			info.SealedDev++
			continue
		}
		ss, err := ReadCoreNLPDoc(path, "draft-"+base[:min(8, len(base))])
		if err != nil {
			info.Unreadable++
			continue
		}
		info.Loaded++
		for _, s := range ss {
			if s == nil {
				continue
			}
			if k := sentenceKeyToks(s.Words); utf8.RuneCountInString(k) >= minSharedRunes && reference[k] {
				info.Deduped++
				continue
			}
			out = append(out, s)
		}
		info.Sentences = len(out)
	}
	return out, info, nil
}

// LoadSilverCorpus loads the CoreNLP-silver part of the Round-2 recipe:
// the eight treatises, ch01..ch11, and the drafts of ext-ref/linear/scratch.
// A draft that shares a sentence (minSharedRunes) with ch13 or ch14 is
// sealed, that is excluded whole, and a draft sentence that already occurs
// in ch01..ch12 or a treatise is dropped. Training code obtains drafts only
// through this function, and the held-out and the development chapters
// enter it as nothing but the sentence keys that exclude their drafts.
func LoadSilverCorpus(root string) (treatises, chapters, drafts []*Sentence, info DraftLoad, err error) {
	treatises, err = LoadSilver(root, TreatiseIDs)
	if err != nil {
		return nil, nil, nil, info, err
	}
	chapters, err = LoadSilver(root, TrainChapters)
	if err != nil {
		return nil, nil, nil, info, err
	}
	seal, err := HoldoutSealKeys(root)
	if err != nil {
		return nil, nil, nil, info, err
	}
	dev, err := DevSealKeys(root)
	if err != nil {
		return nil, nil, nil, info, err
	}
	reference, err := SilverSentenceKeys(root, append(append([]string{}, TreatiseIDs...), BookChapters...))
	if err != nil {
		return nil, nil, nil, info, err
	}
	var di *DraftLoad
	drafts, di, err = LoadSealedDrafts(DraftDir(root), seal, dev, reference)
	if di != nil {
		info = *di
	}
	if err != nil {
		return nil, nil, nil, info, err
	}
	return treatises, chapters, drafts, info, nil
}

// LoadBookDev loads dev_book: the CoreNLP doc of ch12 and its raw text.
// (ch12 is a development set; the held-out chapters are never touched.)
func LoadBookDev(root string) (*corenlp.Doc, string, error) {
	return corenlp.LoadCached(root, DevBookChapter)
}
