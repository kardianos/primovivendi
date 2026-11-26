package parser

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// This file is a small, dependency-free reader/writer for the standard
// HuggingFace .safetensors container (https://github.com/huggingface/safetensors).
// Models trained by this package are stored as flat tensors plus a JSON
// metadata blob, in the same format kardianos/inkline uses.

// Dtypes understood by this codec.
const (
	stF32  = "F32"
	stF16  = "F16"
	stI64  = "I64"
	stI32  = "I32"
	stI16  = "I16"
	stI8   = "I8"
	stU8   = "U8"
	stBool = "BOOL"
)

type stEntry struct {
	DType   string   `json:"dtype"`
	Shape   []int    `json:"shape"`
	Offsets [2]int64 `json:"data_offsets"`
}

// stTensor is one named tensor held in memory by stWriter.
type stTensor struct {
	name  string
	dtype string
	shape []int
	bytes []byte // little-endian payload
}

// stWriter accumulates tensors and writes a .safetensors file.
type stWriter struct {
	tensors []stTensor
	meta    map[string]string
}

func newStWriter() *stWriter { return &stWriter{meta: map[string]string{}} }

// Meta sets one "__metadata__" key/value pair.
func (w *stWriter) Meta(key, val string) { w.meta[key] = val }

func (w *stWriter) add(name, dtype string, shape []int, nbytes int, fill func(dst []byte)) {
	b := make([]byte, nbytes)
	fill(b)
	w.tensors = append(w.tensors, stEntryLike(name, dtype, shape, b))
}

func stEntryLike(name, dtype string, shape []int, b []byte) stTensor {
	return stTensor{name: name, dtype: dtype, shape: shape, bytes: b}
}

// F32 adds a float32 tensor (row-major).
func (w *stWriter) F32(name string, shape []int, data []float32) {
	if numel(shape) != len(data) {
		panic(fmt.Sprintf("safetensors %s: shape %v has %d elements, got %d", name, shape, numel(shape), len(data)))
	}
	w.add(name, stF32, shape, len(data)*4, func(dst []byte) {
		for i, v := range data {
			binary.LittleEndian.PutUint32(dst[i*4:], math.Float32bits(v))
		}
	})
}

// F16 adds an IEEE-754 binary16 tensor converted from float32 (row-major).
func (w *stWriter) F16(name string, shape []int, data []float32) {
	if numel(shape) != len(data) {
		panic(fmt.Sprintf("safetensors %s: shape %v has %d elements, got %d", name, shape, numel(shape), len(data)))
	}
	w.add(name, stF16, shape, len(data)*2, func(dst []byte) {
		for i, v := range data {
			binary.LittleEndian.PutUint16(dst[i*2:], f32ToF16(v))
		}
	})
}

// I64 adds an int64 tensor.
func (w *stWriter) I64(name string, shape []int, data []int64) {
	w.add(name, stI64, shape, len(data)*8, func(dst []byte) {
		for i, v := range data {
			binary.LittleEndian.PutUint64(dst[i*8:], uint64(v))
		}
	})
}

// I32 adds an int32 tensor.
func (w *stWriter) I32(name string, shape []int, data []int32) {
	w.add(name, stI32, shape, len(data)*4, func(dst []byte) {
		for i, v := range data {
			binary.LittleEndian.PutUint32(dst[i*4:], uint32(v))
		}
	})
}

// I16 adds an int16 tensor.
func (w *stWriter) I16(name string, shape []int, data []int16) {
	w.add(name, stI16, shape, len(data)*2, func(dst []byte) {
		for i, v := range data {
			binary.LittleEndian.PutUint16(dst[i*2:], uint16(v))
		}
	})
}

// U8 adds a uint8 tensor.
func (w *stWriter) U8(name string, shape []int, data []byte) {
	w.add(name, stU8, shape, len(data), func(dst []byte) { copy(dst, data) })
}

// JSON adds a UTF-8 JSON document as a U8 tensor named name.
func (w *stWriter) JSON(name string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	w.U8(name, []int{len(b)}, b)
	return nil
}

// Save writes all tensors to path atomically (tmp file + rename).
func (w *stWriter) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	sort.Slice(w.tensors, func(i, j int) bool { return w.tensors[i].name < w.tensors[j].name })
	header := map[string]any{}
	if len(w.meta) > 0 {
		m := map[string]string{}
		for k, v := range w.meta {
			m[k] = v
		}
		header["__metadata__"] = m
	}
	var offset int64
	for _, t := range w.tensors {
		n := int64(len(t.bytes))
		header[t.name] = stEntry{DType: t.dtype, Shape: t.shape, Offsets: [2]int64{offset, offset + n}}
		offset += n
	}
	hdr, err := json.Marshal(header)
	if err != nil {
		return err
	}
	// The official format pads the header with spaces so the data section
	// starts on an 8-byte boundary.
	for (8+len(hdr))%8 != 0 {
		hdr = append(hdr, ' ')
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	var out io.Writer = f
	var gz *gzip.Writer
	if strings.HasSuffix(path, ".gz") {
		gz, err = gzip.NewWriterLevel(f, gzip.BestCompression)
		if err != nil {
			f.Close()
			os.Remove(tmp)
			return err
		}
		out = gz
	}
	if _, err = out.Write(binary.LittleEndian.AppendUint64(nil, uint64(len(hdr)))); err == nil {
		_, err = out.Write(hdr)
	}
	for _, t := range w.tensors {
		if err == nil {
			_, err = out.Write(t.bytes)
		}
	}
	if gz != nil {
		if cerr := gz.Close(); err == nil {
			err = cerr
		}
	}
	if err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

// stFile is an opened .safetensors file.
type stFile struct {
	raw       []byte
	hdr       map[string]stEntry
	meta      map[string]string
	dataStart int64
}

func stOpen(path string) (*stFile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(raw) >= 2 && raw[0] == 0x1f && raw[1] == 0x8b {
		gr, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			return nil, fmt.Errorf("safetensors %s: gzip header: %w", path, err)
		}
		raw, err = io.ReadAll(gr)
		_ = gr.Close()
		if err != nil {
			return nil, fmt.Errorf("safetensors %s: gunzip: %w", path, err)
		}
	}
	if len(raw) < 8 {
		return nil, fmt.Errorf("safetensors %s: truncated header length", path)
	}
	hdrLen := binary.LittleEndian.Uint64(raw[:8])
	if 8+hdrLen > uint64(len(raw)) {
		return nil, fmt.Errorf("safetensors %s: invalid header length %d", path, hdrLen)
	}
	var hdr map[string]json.RawMessage
	if err := json.Unmarshal(raw[8:8+hdrLen], &hdr); err != nil {
		return nil, fmt.Errorf("safetensors %s: bad header: %w", path, err)
	}
	f := &stFile{raw: raw, dataStart: int64(8 + hdrLen), hdr: map[string]stEntry{}, meta: map[string]string{}}
	if m, ok := hdr["__metadata__"]; ok {
		_ = json.Unmarshal(m, &f.meta)
	}
	for name, jm := range hdr {
		if name == "__metadata__" {
			continue
		}
		var e stEntry
		if err := json.Unmarshal(jm, &e); err != nil {
			return nil, fmt.Errorf("safetensors %s: tensor %q: %w", path, name, err)
		}
		if e.Offsets[0] < 0 || e.Offsets[1] < e.Offsets[0] || f.dataStart+e.Offsets[1] > int64(len(raw)) {
			return nil, fmt.Errorf("safetensors %s: bad offsets for %q", path, name)
		}
		f.hdr[name] = e
	}
	return f, nil
}

// Has reports whether a tensor with this name exists.
func (f *stFile) Has(name string) bool { _, ok := f.hdr[name]; return ok }

// Meta returns a "__metadata__" value ("" when absent).
func (f *stFile) Meta(key string) string { return f.meta[key] }

func (f *stFile) bytes(name, dtype string, item int) ([]byte, []int, error) {
	e, ok := f.hdr[name]
	if !ok {
		return nil, nil, fmt.Errorf("missing tensor %q", name)
	}
	if e.DType != dtype {
		return nil, nil, fmt.Errorf("tensor %q has dtype %s, want %s", name, e.DType, dtype)
	}
	n := e.Offsets[1] - e.Offsets[0]
	if n%int64(item) != 0 || int64(numel(e.Shape))*int64(item) != n {
		return nil, nil, fmt.Errorf("tensor %q: shape %v does not match %d bytes", name, e.Shape, n)
	}
	a := f.dataStart + e.Offsets[0]
	return f.raw[a : a+n], e.Shape, nil
}

// F32 returns a float32 tensor and its shape, transparently promoting F16
// tensors when the model was stored in half precision.
func (f *stFile) F32(name string) ([]float32, []int, error) {
	if e, ok := f.hdr[name]; ok && e.DType == stF16 {
		b, shape, err := f.bytes(name, stF16, 2)
		if err != nil {
			return nil, nil, err
		}
		out := make([]float32, len(b)/2)
		for i := range out {
			out[i] = f16ToF32(binary.LittleEndian.Uint16(b[i*2:]))
		}
		return out, shape, nil
	}
	b, shape, err := f.bytes(name, stF32, 4)
	if err != nil {
		return nil, nil, err
	}
	out := make([]float32, len(b)/4)
	for i := range out {
		out[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
	}
	return out, shape, nil
}

// PackSafetensors copies a .safetensors file to dstPath (gzip-compressed when
// dstPath ends with ".gz"), optionally quantizing F32 tensors to IEEE-754 F16.
func PackSafetensors(srcPath, dstPath string, useF16 bool) error {
	f, err := stOpen(srcPath)
	if err != nil {
		return err
	}
	w := newStWriter()
	for k, v := range f.meta {
		w.Meta(k, v)
	}
	if useF16 {
		w.Meta("packed_dtype", stF16)
	}
	names := make([]string, 0, len(f.hdr))
	for name := range f.hdr {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		e := f.hdr[name]
		if useF16 && e.DType == stF32 {
			vals, shape, err := f.F32(name)
			if err != nil {
				return err
			}
			w.F16(name, shape, vals)
			continue
		}
		a := f.dataStart + e.Offsets[0]
		n := e.Offsets[1] - e.Offsets[0]
		buf := make([]byte, n)
		copy(buf, f.raw[a:a+n])
		w.tensors = append(w.tensors, stTensor{name: name, dtype: e.DType, shape: e.Shape, bytes: buf})
	}
	return w.Save(dstPath)
}

// f32ToF16 converts a float32 to IEEE-754 binary16 with round-to-nearest-even.
func f32ToF16(v float32) uint16 {
	bits := math.Float32bits(v)
	sign := uint16((bits >> 16) & 0x8000)
	exp := int((bits>>23)&0xFF) - 127 + 15
	mant := bits & 0x7FFFFF
	if exp <= 0 {
		if exp < -10 {
			return sign
		}
		mant = (mant | 0x800000) >> uint(1-exp)
		if (mant & 0x1000) != 0 {
			mant += 0x2000
		}
		return sign | uint16(mant>>13)
	}
	if exp >= 31 {
		if exp == 128+15 && mant != 0 {
			return sign | 0x7E00 // NaN
		}
		return sign | 0x7C00 // Inf
	}
	// Round-to-nearest-even.
	lsb := (mant >> 13) & 1
	mant += 0x0FFF + lsb
	if (mant & 0x800000) != 0 {
		mant = 0
		exp++
		if exp >= 31 {
			return sign | 0x7C00
		}
	}
	return sign | uint16(exp<<10) | uint16(mant>>13)
}

// f16ToF32 converts an IEEE-754 binary16 bit pattern to float32.
func f16ToF32(h uint16) float32 {
	sign := uint32(h&0x8000) << 16
	exp := uint32((h >> 10) & 0x1F)
	mant := uint32(h & 0x03FF)
	switch exp {
	case 0:
		if mant == 0 {
			return math.Float32frombits(sign)
		}
		// Subnormal.
		exp = 127 - 14
		for (mant & 0x0400) == 0 {
			mant <<= 1
			exp--
		}
		mant &= 0x03FF
		return math.Float32frombits(sign | (exp << 23) | (mant << 13))
	case 31:
		if mant == 0 {
			return math.Float32frombits(sign | 0x7F800000)
		}
		return math.Float32frombits(sign | 0x7FC00000 | (mant << 13))
	default:
		return math.Float32frombits(sign | ((exp + 127 - 15) << 23) | (mant << 13))
	}
}

// I64 returns an int64 tensor and its shape.
func (f *stFile) I64(name string) ([]int64, []int, error) {
	b, shape, err := f.bytes(name, stI64, 8)
	if err != nil {
		return nil, nil, err
	}
	out := make([]int64, len(b)/8)
	for i := range out {
		out[i] = int64(binary.LittleEndian.Uint64(b[i*8:]))
	}
	return out, shape, nil
}

// I32 returns an int32 tensor and its shape.
func (f *stFile) I32(name string) ([]int32, []int, error) {
	b, shape, err := f.bytes(name, stI32, 4)
	if err != nil {
		return nil, nil, err
	}
	out := make([]int32, len(b)/4)
	for i := range out {
		out[i] = int32(binary.LittleEndian.Uint32(b[i*4:]))
	}
	return out, shape, nil
}

// I16 returns an int16 tensor and its shape.
func (f *stFile) I16(name string) ([]int16, []int, error) {
	b, shape, err := f.bytes(name, stI16, 2)
	if err != nil {
		return nil, nil, err
	}
	out := make([]int16, len(b)/2)
	for i := range out {
		out[i] = int16(binary.LittleEndian.Uint16(b[i*2:]))
	}
	return out, shape, nil
}

// U8 returns a uint8 tensor and its shape.
func (f *stFile) U8(name string) ([]byte, []int, error) {
	return f.bytes(name, stU8, 1)
}

// JSON decodes a U8 tensor holding a JSON document into v.
func (f *stFile) JSON(name string, v any) error {
	b, _, err := f.U8(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func numel(shape []int) int {
	n := 1
	for _, d := range shape {
		n *= d
	}
	return n
}
