// mst.go — the decoder of the biaffine parser: the maximum spanning
// arborescence of Chu & Liu / Edmonds over the complete digraph of the N+1
// nodes (ROOT is node 0), constrained to a single child of ROOT. The
// constraint always holds exactly; which trees are compared for it is the
// candidate rule of mst.
package parser

import "sort"

// mstCandidateRoots is how many tokens are tried as the root, by root
// margin, when Chu-Liu/Edmonds returns several.
const mstCandidateRoots = 8

// mstNoEdge stands for an absent edge: every real arc score is far above
// it, and unlike −Inf it stays finite under subtraction.
const mstNoEdge = float32(-1e30)

// mst fills heads[1..N] with the governors of the best-scoring single-root
// tree over the arc scores s that the candidate rule below reaches: s is N
// rows of N+1 scores, row i−1 holding the score of dependent i under every
// head j = 0..N, with s[i−1][i] at −Inf.
//
// Chu-Liu/Edmonds over the unconstrained scores already returns a single
// child of ROOT for 99% of sentences (17 of 2,001 on UD-EWT dev), and that
// tree is the exact optimum. Otherwise candidate roots r are tried by
// forbidding all other root arcs and the best valid tree wins: the eight
// tokens whose root arc loses least against their best other head
// (S[r][0] − max_{j>0} S[r][j]) and the roots already chosen, which the
// specification allows in place of every token. The two rules gave the
// same 2,001 trees on UD-EWT dev.
func mst(s []float32, N int, heads []int32) {
	if N <= 0 {
		return
	}
	W := N + 1
	w := make([]float32, W*W)
	for i := range w {
		w[i] = mstNoEdge // no edge enters ROOT
	}
	for i := 1; i <= N; i++ {
		for j := 0; j <= N; j++ {
			if v := s[(i-1)*W+j]; v > mstNoEdge {
				w[i*W+j] = v
			}
		}
		w[i*W+i] = mstNoEdge
	}
	h := cle(w, W)
	nRoots := 0
	for i := 1; i <= N; i++ {
		if h[i] == 0 {
			nRoots++
		}
	}
	if nRoots == 1 {
		copy(heads[1:N+1], h[1:])
		return
	}
	type candidate struct {
		r      int
		margin float32 // w[i][0] − max_{j>0} w[i][j]
	}
	cands := make([]candidate, 0, N)
	for i := 1; i <= N; i++ {
		best := mstNoEdge
		for j := 1; j <= N; j++ {
			if j != i && w[i*W+j] > best {
				best = w[i*W+j]
			}
		}
		cands = append(cands, candidate{i, w[i*W] - best})
	}
	sort.Slice(cands, func(a, b int) bool {
		if cands[a].margin != cands[b].margin {
			return cands[a].margin > cands[b].margin
		}
		return cands[a].r < cands[b].r
	})
	saved := make([]float32, N)
	tried := make([]bool, N+1)
	best, bestScore := h, float32(mstNoEdge)
	try := func(r int) {
		if tried[r] {
			return
		}
		tried[r] = true
		for i := 1; i <= N; i++ {
			saved[i-1] = w[i*W]
			if i != r {
				w[i*W] = mstNoEdge
			}
		}
		hr := cle(w, W)
		for i := 1; i <= N; i++ {
			w[i*W] = saved[i-1]
		}
		if hr[r] != 0 || !validTree(hr, N) {
			return // ROOT does not have r as its only child, or a cycle is left
		}
		score := float32(0)
		for i := 1; i <= N; i++ {
			score += w[i*W+int(hr[i])]
		}
		if score > bestScore {
			best, bestScore = hr, score
		}
	}
	for _, c := range cands[:min(len(cands), mstCandidateRoots)] {
		try(c.r)
	}
	for i := 1; i <= N; i++ {
		if h[i] == 0 {
			try(i)
		}
	}
	if bestScore > mstNoEdge {
		copy(heads[1:N+1], best[1:])
		return
	}
	// No candidate gives a valid tree: no token prefers ROOT to another
	// head. The token with the best root arc becomes the child of ROOT
	// regardless, and the rest hang from ROOT or from it.
	r := 1
	for i := 2; i <= N; i++ {
		if w[i*W] > w[r*W] {
			r = i
		}
	}
	id := make([]int32, W) // token -> node of the forced problem
	inv := []int32{int32(r)}
	for i := 1; i <= N; i++ {
		if i == r {
			id[i] = 0
		} else {
			id[i] = int32(len(inv))
			inv = append(inv, int32(i))
		}
	}
	m := len(inv)
	u := make([]float32, m*m)
	for i := range u {
		u[i] = mstNoEdge
	}
	for i := 1; i <= N; i++ {
		if i == r {
			continue
		}
		for j := 0; j <= N; j++ {
			if j == 0 || j == r { // ROOT and r are one node here
				u[int(id[i])*m] = max(u[int(id[i])*m], w[i*W+j])
			} else if j != i {
				u[int(id[i])*m+int(id[j])] = w[i*W+j]
			}
		}
	}
	hf := cle(u, m)
	heads[r] = 0
	for i := 1; i <= N; i++ {
		if i != r {
			heads[i] = inv[hf[id[i]]]
		}
	}
}

// cle returns a maximum spanning arborescence of the graph w rooted at node
// 0: w is n×n row-major, w[i·n+j] the score of edge j → i for i ≥ 1 (an
// edge into node 0, a self edge, or a score at most mstNoEdge does not
// exist). The result holds the parent of every node i ≥ 1, 0 for the
// children of the root. Cycles are contracted and the smaller graph solved
// the same way.
func cle(w []float32, n int) []int32 {
	parent := make([]int32, n)
	for i := 1; i < n; i++ {
		best := mstNoEdge
		for j := 0; j < n; j++ {
			if v := w[i*n+j]; j != i && v > best {
				best, parent[i] = v, int32(j)
			}
		}
	}
	// The first cycle among the best edges.
	state := make([]int8, n) // 0 unseen, 1 on the path being walked, 2 closed
	var cycle []int32
	for i := 1; i < n && cycle == nil; i++ {
		if state[i] != 0 {
			continue
		}
		// Walk the best edges from i until reaching a closed node, a node
		// of the path being walked (a cycle), or the root.
		path := make([]int32, 0, 8)
		u := int32(i)
		for u != 0 && state[u] == 0 {
			state[u] = 1
			path = append(path, u)
			u = parent[u]
		}
		if u != 0 && state[u] == 1 {
			k := 0
			for path[k] != u {
				k++
			}
			cycle = append(cycle, path[k:]...)
		}
		for _, x := range path {
			state[x] = 2
		}
	}
	if cycle == nil {
		return parent
	}
	// Contract the cycle into one node (its first member).
	inCycle := make([]bool, n)
	for _, v := range cycle {
		inCycle[v] = true
	}
	c := cycle[0]
	id := make([]int32, n) // old node -> contracted node (−1 inside the cycle)
	inv := []int32{0, c}
	id[0], id[c] = 0, 1
	for v := 1; v < n; v++ {
		if !inCycle[v] {
			id[v] = int32(len(inv))
			inv = append(inv, int32(v))
		}
	}
	m := len(inv)
	u := make([]float32, m*m)
	for i := range u {
		u[i] = mstNoEdge
	}
	for i := 1; i < n; i++ {
		if inCycle[i] {
			continue
		}
		for j := 0; j < n; j++ {
			if j != i && !inCycle[j] {
				u[int(id[i])*m+int(id[j])] = w[i*n+j]
			}
		}
	}
	// An edge u → c replaces the cycle edge into its endpoint v, so it
	// scores w[v][u] − w[v][parent[v]] for the best v of the cycle (kept in
	// inChoice[u]); an edge c → x scores the best w[x][v] over the nodes v
	// of the cycle (kept in outChoice[x]).
	inChoice, outChoice := make([]int32, m), make([]int32, m)
	for i := 0; i < n; i++ {
		if inCycle[i] {
			continue
		}
		best := 2 * mstNoEdge
		for _, v := range cycle {
			if x := w[int(v)*n+i] - w[int(v)*n+int(parent[v])]; x > best {
				best, inChoice[id[i]] = x, v
			}
		}
		if best > mstNoEdge {
			u[int(id[c])*m+int(id[i])] = best
		}
	}
	for x := 1; x < n; x++ {
		if inCycle[x] {
			continue
		}
		best := mstNoEdge
		for _, v := range cycle {
			if w[x*n+int(v)] > best {
				best, outChoice[id[x]] = w[x*n+int(v)], v
			}
		}
		u[int(id[x])*m+int(id[c])] = best
	}
	h := cle(u, m)
	head := make([]int32, n)
	for i := 1; i < n; i++ {
		switch {
		case inCycle[i]:
			head[i] = parent[i]
		case h[id[i]] == id[c]:
			head[i] = outChoice[id[i]]
		default:
			head[i] = inv[h[id[i]]]
		}
	}
	// The cycle hangs from its contracted node's parent q: q enters it at
	// the node whose cycle edge is replaced.
	head[inChoice[h[id[c]]]] = inv[h[id[c]]]
	return head
}

// validTree reports whether head[1..n] (1-based governors, 0 = ROOT) is an
// acyclic tree in which every token reaches ROOT.
func validTree(head []int32, n int) bool {
	for i := 1; i <= n; i++ {
		cur, steps := int32(i), 0
		for cur != 0 {
			h := int32(0)
			if int(cur) < len(head) {
				h = head[cur]
			}
			if h < 0 || h > int32(n) || h == cur {
				return false
			}
			cur = h
			if steps++; steps > n {
				return false
			}
		}
	}
	return true
}
