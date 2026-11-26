// oracle.go — Stage 3: the Arc-Standard transition system (Nivre 2004): the
// configuration, its legal-action rules, the static oracle that turns a gold
// tree into a transition sequence, and the 48 fixed feature slots of Chen &
// Manning (2014) that the neural parser consumes.
package parser

// Transition layout: SHIFT; LEFT-ARC(r) and RIGHT-ARC(r) for relation numbers
// r in [0, nDepRels), where r indexes DepLabels[1:] (the non-ROOT labels); and
// RIGHT-ARC(ROOT), the last action. The ROOT arc is its own transition, legal
// only when the stack is [.., ROOT, x] and the buffer is empty, so any run
// yields exactly one root.
const actShift = 0

var (
	nDepRels = len(DepLabels) - 1 // non-ROOT labels
	nDepActs = 2*nDepRels + 2     // SHIFT + LEFT-ARC + RIGHT-ARC + RIGHT-ARC(ROOT)
	depRoot  = nDepActs - 1       // RIGHT-ARC(ROOT)
)

func actLeft(r int) int  { return 1 + r }
func actRight(r int) int { return 1 + nDepRels + r }

// depActLabel returns the DepLabels index of the label an arc action assigns.
func depActLabel(a int) int {
	switch {
	case a == depRoot:
		return 0
	case a <= nDepRels:
		return a
	}
	return a - nDepRels
}

// actForLabel returns LEFT-ARC's action number for a non-ROOT DepLabels index.
func relNum(label int) int { return label - 1 }

// Feature slots (Chen & Manning 2014): 18 token positions, their 18 POS tags,
// and the 12 dependency labels of the child/grandchild positions.
const (
	nWordSlots = 18
	nDepSlots  = 12
	nFeatSlots = nWordSlots + nWordSlots + nDepSlots
)

// config is an Arc-Standard configuration over one sentence. Token numbers
// are 1-based; 0 is the synthetic ROOT, which always sits at the bottom of the
// stack. The buffer is [buf, n]; buf == n+1 means it is empty.
type config struct {
	n     int32
	stack []int32 // top is the last element
	buf   int32
	// Arcs assigned so far (-1 while unset).
	head []int32
	rel  []int32
	// First and second dependents attached on each side, and their labels.
	lc1, lc2, rc1, rc2 []int32
	ll1, ll2, rl1, rl2 []int32
}

// depScratch is the reusable per-call working memory of the parser.
type depScratch struct {
	c         config
	stack     []int32
	head, rel []int32
	lc1, lc2  []int32
	rc1, rc2  []int32
	ll1, ll2  []int32
	rl1, rl2  []int32
	mask      []bool
	logits    []float32 // action scores of one configuration
	scratch   []float32 // embedding-sized scratch for uncached words
	hidden    []float32 // hidden layer of the parser (sized by ensureHidden)
	goldHead  []int32
	goldRel   []int32
	left      []int32
	forced    int
	tok       [nWordSlots]int32
	lab       [nDepSlots]int32
}

func newDepScratch() *depScratch {
	return &depScratch{mask: make([]bool, nDepActs), logits: make([]float32, nDepActs), scratch: make([]float32, maxEmbed)}
}

// maxEmbed bounds the word embedding width (Dw) the scratch supports.
const maxEmbed = 128

// ensureHidden makes room for an H-unit hidden layer.
func (s *depScratch) ensureHidden(H int) {
	if cap(s.hidden) < H {
		s.hidden = make([]float32, H)
	}
	if cap(s.logits) < nDepActs {
		s.logits = make([]float32, nDepActs)
	}
	if cap(s.scratch) < maxEmbed {
		s.scratch = make([]float32, maxEmbed)
	}
}

// reset starts a fresh configuration over an n-token sentence, growing the
// scratch buffers when needed.
func (s *depScratch) reset(n int) {
	if cap(s.head) < n+1 {
		for _, p := range []*[]int32{&s.head, &s.rel, &s.lc1, &s.lc2, &s.rc1, &s.rc2, &s.ll1, &s.ll2, &s.rl1, &s.rl2, &s.goldHead, &s.goldRel, &s.left} {
			*p = make([]int32, n+1, max(n+1, 256))
		}
		s.stack = make([]int32, 0, n+1)
	}
	c := &s.c
	c.n = int32(n)
	c.stack = append(s.stack[:0], 0)
	c.buf = 1
	c.head = s.head[:n+1]
	c.rel = s.rel[:n+1]
	c.lc1, c.lc2, c.rc1, c.rc2 = s.lc1[:n+1], s.lc2[:n+1], s.rc1[:n+1], s.rc2[:n+1]
	c.ll1, c.ll2, c.rl1, c.rl2 = s.ll1[:n+1], s.ll2[:n+1], s.rl1[:n+1], s.rl2[:n+1]
	for i := 1; i <= n; i++ {
		c.head[i] = -1
	}
	clear(c.rel[1 : n+1])
	clear(c.lc1[1 : n+1])
	clear(c.lc2[1 : n+1])
	clear(c.rc1[1 : n+1])
	clear(c.rc2[1 : n+1])
	clear(c.ll1[1 : n+1])
	clear(c.ll2[1 : n+1])
	clear(c.rl1[1 : n+1])
	clear(c.rl2[1 : n+1])
	c.lc1[0], c.lc2[0], c.rc1[0], c.rc2[0] = 0, 0, 0, 0
	c.ll1[0], c.ll2[0], c.rl1[0], c.rl2[0] = 0, 0, 0, 0
}

func (c *config) terminal() bool { return c.buf > c.n && len(c.stack) <= 1 }

// top returns the k-th stack item from the top (0 = top), or 0 when absent.
func (c *config) top(k int) int32 {
	if len(c.stack) > k {
		return c.stack[len(c.stack)-1-k]
	}
	return 0
}

// legal marks the transitions that can be applied in the current state.
func (c *config) legal(mask []bool) {
	clear(mask)
	if c.buf <= c.n {
		mask[actShift] = true
	}
	if len(c.stack) >= 2 {
		if c.stack[len(c.stack)-2] != 0 {
			for a := 1; a < depRoot; a++ {
				mask[a] = true
			}
		} else if c.buf > c.n {
			mask[depRoot] = true
		}
	}
}

// canApply reports whether action a is legal here.
func (c *config) canApply(a int) bool {
	if a == actShift {
		return c.buf <= c.n
	}
	if len(c.stack) < 2 {
		return false
	}
	if c.stack[len(c.stack)-2] != 0 {
		return a >= 1 && a < depRoot
	}
	return a == depRoot && c.buf > c.n
}

// attach records that d is a dependent of h with DepLabels index lab.
func (c *config) attach(h, d, lab int32) {
	c.head[d] = h
	c.rel[d] = lab
	if h == 0 { // the synthetic ROOT keeps no child features
		return
	}
	if d < h {
		c.lc2[h], c.ll2[h] = c.lc1[h], c.ll1[h]
		c.lc1[h], c.ll1[h] = d, lab
	} else {
		c.rc2[h], c.rl2[h] = c.rc1[h], c.rl1[h]
		c.rc1[h], c.rl1[h] = d, lab
	}
}

// apply executes action a and reports false when it is illegal there.
func (c *config) apply(a int) bool {
	if !c.canApply(a) {
		return false
	}
	switch {
	case a == actShift:
		c.stack = append(c.stack, c.buf)
		c.buf++
	case a <= nDepRels: // LEFT-ARC: head = s0, dependent = s1 (s1 is removed)
		s0, s1 := c.top(0), c.top(1)
		c.attach(s0, s1, int32(depActLabel(a)))
		c.stack[len(c.stack)-2] = s0
		c.stack = c.stack[:len(c.stack)-1]
	case a < depRoot: // RIGHT-ARC: head = s1, dependent = s0
		s0, s1 := c.top(0), c.top(1)
		c.attach(s1, s0, int32(depActLabel(a)))
		c.stack = c.stack[:len(c.stack)-1]
	default: // RIGHT-ARC(ROOT)
		c.attach(0, c.top(0), 0)
		c.stack = c.stack[:len(c.stack)-1]
	}
	return true
}

// features returns the token at each of the 18 positions (0 = absent) and the
// dependency label at each of the 12 child/grandchild positions (0 = none).
func (c *config) features(tok *[nWordSlots]int32, lab *[nDepSlots]int32) {
	s0, s1 := c.top(0), c.top(1)
	tok[0], tok[1], tok[2] = s0, s1, c.top(2)
	for k := int32(0); k < 3; k++ {
		if c.buf+k <= c.n {
			tok[3+k] = c.buf + k
		} else {
			tok[3+k] = 0
		}
	}
	tok[6], tok[7], tok[8], tok[9] = c.lc1[s0], c.rc1[s0], c.lc2[s0], c.rc2[s0]
	tok[10], tok[11], tok[12], tok[13] = c.lc1[s1], c.rc1[s1], c.lc2[s1], c.rc2[s1]
	lab[0], lab[1], lab[2], lab[3] = c.ll1[s0], c.rl1[s0], c.ll2[s0], c.rl2[s0]
	lab[4], lab[5], lab[6], lab[7] = c.ll1[s1], c.rl1[s1], c.ll2[s1], c.rl2[s1]
	g := c.lc1[s0]
	tok[14], lab[8] = c.lc1[g], c.ll1[g]
	g = c.rc1[s0]
	tok[15], lab[9] = c.rc1[g], c.rl1[g]
	g = c.lc1[s1]
	tok[16], lab[10] = c.lc1[g], c.ll1[g]
	g = c.rc1[s1]
	tok[17], lab[11] = c.rc1[g], c.rl1[g]
}

// ---- static oracle -----------------------------------------------------------

// isProjective reports whether heads (1-based, head[i] = governor, 0 = root)
// is a single-root, acyclic, projective tree.
func isProjective(heads []int32) bool {
	n := len(heads) - 1
	if n == 0 {
		return true
	}
	roots := 0
	for i := 1; i <= n; i++ {
		if heads[i] == 0 {
			roots++
		} else if heads[i] < 0 || heads[i] > int32(n) || heads[i] == int32(i) {
			return false
		}
	}
	if roots != 1 {
		return false
	}
	for i := int32(1); i <= int32(n); i++ {
		cur, steps := i, 0
		for heads[cur] != 0 {
			cur = heads[cur]
			if steps++; steps > n {
				return false
			}
		}
	}
	// Projectivity: every token between a dependent and its head descends
	// from the head, and nothing to the left of the root escapes it.
	for d := int32(1); d <= int32(n); d++ {
		h := heads[d]
		lo, hi := d, h
		if lo > hi {
			lo, hi = hi, lo
		}
		for m := lo + 1; m < hi; m++ {
			if !descends(heads, m, h) {
				return false
			}
		}
	}
	for i := int32(1); i <= int32(n); i++ {
		if heads[i] == 0 {
			for m := int32(1); m < i; m++ {
				if !descends(heads, m, i) {
					return false
				}
			}
			break
		}
	}
	return true
}

func descends(heads []int32, m, h int32) bool {
	steps := 0
	for m != 0 {
		if m == h {
			return true
		}
		m = heads[m]
		if steps++; steps > len(heads) {
			return false
		}
	}
	return h == 0
}

// oracleStep returns the gold action for the configuration and the head token
// of the arc it creates (-1 for SHIFT), or (-1, -1) when the static oracle is
// stuck on a non-projective gold tree.
func (s *depScratch) oracleStep() (act, head int) {
	c := &s.c
	s0, s1 := c.top(0), c.top(1)
	if len(c.stack) >= 2 && s1 != 0 {
		if s.goldHead[s1] == s0 && s.left[s1] == 0 {
			return actLeft(relNum(int(s.goldRel[s1]))), int(s0)
		}
		if s.goldHead[s0] == s1 && s.left[s0] == 0 {
			return actRight(relNum(int(s.goldRel[s0]))), int(s1)
		}
	} else if len(c.stack) >= 2 && s.goldHead[s0] == 0 && s.left[s0] == 0 && c.buf > c.n {
		return depRoot, -1
	}
	if c.buf <= c.n {
		return actShift, -1
	}
	return -1, -1
}

// oracleSequence runs the static oracle over a gold tree (heads 1-based with
// a single root; rels are DepLabels indices), appending every step's feature
// tokens (18 uint16), labels (12 uint8), and gold action to tok, lab, and act.
// It reports whether the sequence reproduced the tree exactly (no forced
// transition): that is always the case for a projective tree.
func (s *depScratch) oracleSequence(heads, rels []int32, tok []uint16, lab []uint8, act []int16) ([]uint16, []uint8, []int16, bool) {
	n := len(heads) - 1
	s.reset(n)
	s.goldHead = append(s.goldHead[:0], heads...)
	s.goldRel = append(s.goldRel[:0], rels...)
	s.left = s.left[:n+1]
	clear(s.left)
	for d := 1; d <= n; d++ {
		s.left[heads[d]]++
	}
	exact := true
	for !s.c.terminal() {
		s.c.features(&s.tok, &s.lab)
		a, head := s.oracleStep()
		if a < 0 {
			// A non-projective gold tree: force a legal arc so that every
			// sentence still takes exactly 2N transitions.
			if s.c.top(1) != 0 {
				a = actLeft(relNum(depIndex["dep"]))
			} else {
				a = depRoot
			}
			exact = false
		}
		for i := 0; i < nWordSlots; i++ {
			tok = append(tok, uint16(s.tok[i]))
		}
		for i := 0; i < nDepSlots; i++ {
			lab = append(lab, uint8(s.lab[i]))
		}
		act = append(act, int16(a))
		s.c.apply(a)
		if head >= 0 && s.left[head] > 0 {
			s.left[head]--
		}
	}
	return tok, lab, act, exact
}
