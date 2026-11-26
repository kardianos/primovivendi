package vec

import "math/bits"

// WordFilter is a stack-allocated, zero-GC 256-bit (4 x uint64) Bloom filter
// for O(1) bitwise word membership, echo pre-screening, and glossary lock
// pre-filtering without map lookups or heap allocations.
type WordFilter [4]uint64

// HashWord computes a fast case-insensitive FNV-1a 64-bit hash of an ASCII/UTF-8 word.
func HashWord(s string) uint64 {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)
	h := uint64(offset64)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		h ^= uint64(c)
		h *= prime64
	}
	return h
}

// HashWordBytes hashes a lowercase/ASCII byte slice with zero allocations.
func HashWordBytes(b []byte) uint64 {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)
	h := uint64(offset64)
	for _, c := range b {
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		h ^= uint64(c)
		h *= prime64
	}
	return h
}

// HashBigram packs two 64-bit word hashes into one collision-resistant 64-bit
// bigram key without concatenating strings.
func HashBigram(h1, h2 uint64) uint64 {
	x := h1 ^ bits.RotateLeft64(h2, 29) ^ 0x9e3779b97f4a7c15
	x ^= x >> 33
	x *= 0xff51afd7ed558ccd
	x ^= x >> 33
	x *= 0xc4ceb9fe1a85ec53
	x ^= x >> 33
	return x
}

// Add inserts a word into the 256-bit filter using 3 independent bit positions.
func (f *WordFilter) Add(word string) {
	f.AddHash(HashWord(word))
}

// AddHash inserts a precomputed 64-bit word hash into the 256-bit filter.
func (f *WordFilter) AddHash(h uint64) {
	b1 := h & 255
	b2 := (h >> 8) & 255
	b3 := (h >> 16) & 255
	f[b1>>6] |= 1 << (b1 & 63)
	f[b2>>6] |= 1 << (b2 & 63)
	f[b3>>6] |= 1 << (b3 & 63)
}

// MaybeHas reports whether word may be in the filter (zero false negatives).
func (f *WordFilter) MaybeHas(word string) bool {
	return f.MaybeHasHash(HashWord(word))
}

// MaybeHasHash checks membership of a precomputed hash in ~3 CPU instructions.
func (f *WordFilter) MaybeHasHash(h uint64) bool {
	b1 := h & 255
	b2 := (h >> 8) & 255
	b3 := (h >> 16) & 255
	return (f[b1>>6]&(1<<(b1&63))) != 0 &&
		(f[b2>>6]&(1<<(b2&63))) != 0 &&
		(f[b3>>6]&(1<<(b3&63))) != 0
}

// OverlapBits returns the bitwise POPCNT of shared set bits between two filters.
func (f *WordFilter) OverlapBits(o *WordFilter) int {
	return bits.OnesCount64(f[0]&o[0]) +
		bits.OnesCount64(f[1]&o[1]) +
		bits.OnesCount64(f[2]&o[2]) +
		bits.OnesCount64(f[3]&o[3])
}

// BitFilter is a fixed-size power-of-two bitwise Bloom filter (e.g. 64KB = 524,288 bits)
// used to reject unseen reference-corpus bigrams in O(1) without touching a Go map.
type BitFilter struct {
	words []uint64
	mask  uint64
}

// NewBitFilter allocates a Bloom filter with 1<<pow2Bits bits (minimum 1<<10).
func NewBitFilter(pow2Bits uint) *BitFilter {
	if pow2Bits < 10 {
		pow2Bits = 10
	}
	nBits := uint64(1) << pow2Bits
	return &BitFilter{
		words: make([]uint64, nBits/64),
		mask:  nBits - 1,
	}
}

// Add sets 2 bit positions for hash h.
func (bf *BitFilter) Add(h uint64) {
	if bf == nil || len(bf.words) == 0 {
		return
	}
	i1 := h & bf.mask
	i2 := (h >> 21) & bf.mask
	bf.words[i1>>6] |= 1 << (i1 & 63)
	bf.words[i2>>6] |= 1 << (i2 & 63)
}

// MaybeHas checks 2 bit positions for hash h (zero false negatives).
func (bf *BitFilter) MaybeHas(h uint64) bool {
	if bf == nil || len(bf.words) == 0 {
		return false
	}
	i1 := h & bf.mask
	i2 := (h >> 21) & bf.mask
	return (bf.words[i1>>6]&(1<<(i1&63))) != 0 &&
		(bf.words[i2>>6]&(1<<(i2&63))) != 0
}
