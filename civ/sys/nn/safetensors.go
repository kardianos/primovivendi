package nn

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
)

type stEntry struct {
	DType   string   `json:"dtype"`
	Shape   []int    `json:"shape"`
	Offsets [2]int64 `json:"data_offsets"`
}

// SaveSafetensors writes all Ranker weights and normalization buffers in
// standard HuggingFace .safetensors format (just like kardianos/inkline).
func (r *Ranker) SaveSafetensors(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	type named struct {
		name  string
		shape []int
		data  []float32
	}
	tensors := []named{
		{"norm.mean", []int{r.InDim}, r.Mean},
		{"norm.inv_std", []int{r.InDim}, r.InvStd},
		{"l1.weight", []int{r.L1.Out, r.L1.In}, r.L1.W},
		{"l1.bias", []int{r.L1.Out}, r.L1.B},
		{"l2.weight", []int{r.L2.Out, r.L2.In}, r.L2.W},
		{"l2.bias", []int{r.L2.Out}, r.L2.B},
		{"l3.weight", []int{r.L3.Out, r.L3.In}, r.L3.W},
		{"l3.bias", []int{r.L3.Out}, r.L3.B},
		{"skip.weight", []int{r.Skip.Out, r.Skip.In}, r.Skip.W},
		{"skip.bias", []int{r.Skip.Out}, r.Skip.B},
	}

	header := map[string]any{
		"__metadata__": map[string]string{
			"format":  "civ-compose-ranker-v1",
			"trained": strconv.Itoa(r.Trained),
			"step":    strconv.Itoa(r.Step),
		},
	}
	var offset int64
	for _, t := range tensors {
		nBytes := int64(len(t.data) * 4)
		header[t.name] = stEntry{
			DType:   "F32",
			Shape:   t.shape,
			Offsets: [2]int64{offset, offset + nBytes},
		}
		offset += nBytes
	}
	hdrBytes, err := json.Marshal(header)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint64(len(hdrBytes))); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if _, err := f.Write(hdrBytes); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	buf := make([]byte, 4)
	for _, t := range tensors {
		for _, v := range t.data {
			binary.LittleEndian.PutUint32(buf, math.Float32bits(v))
			if _, err := f.Write(buf); err != nil {
				f.Close()
				os.Remove(tmp)
				return err
			}
		}
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

// LoadSafetensors loads a Ranker saved by SaveSafetensors.
func LoadSafetensors(path string) (*Ranker, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(raw) < 8 {
		return nil, fmt.Errorf("safetensors %s: truncated header length", path)
	}
	hdrLen := binary.LittleEndian.Uint64(raw[:8])
	if int64(8+hdrLen) > int64(len(raw)) {
		return nil, fmt.Errorf("safetensors %s: invalid header length %d", path, hdrLen)
	}
	var hdr map[string]json.RawMessage
	if err := json.Unmarshal(raw[8:8+hdrLen], &hdr); err != nil {
		return nil, err
	}
	dataStart := int64(8 + hdrLen)
	readF32 := func(name string) ([]float32, []int, error) {
		jm, ok := hdr[name]
		if !ok {
			return nil, nil, fmt.Errorf("missing tensor %q", name)
		}
		var e stEntry
		if err := json.Unmarshal(jm, &e); err != nil {
			return nil, nil, err
		}
		a, b := dataStart+e.Offsets[0], dataStart+e.Offsets[1]
		if a < dataStart || b > int64(len(raw)) || (b-a)%4 != 0 {
			return nil, nil, fmt.Errorf("bad offsets for %q", name)
		}
		n := int((b - a) / 4)
		out := make([]float32, n)
		for i := 0; i < n; i++ {
			bits := binary.LittleEndian.Uint32(raw[a+int64(i*4) : a+int64(i*4)+4])
			out[i] = math.Float32frombits(bits)
		}
		return out, e.Shape, nil
	}

	r := NewRanker(1)
	var err2 error
	if r.Mean, _, err2 = readF32("norm.mean"); err2 != nil {
		return nil, err2
	}
	if r.InvStd, _, err2 = readF32("norm.inv_std"); err2 != nil {
		return nil, err2
	}
	var sh []int
	if r.L1.W, sh, err2 = readF32("l1.weight"); err2 != nil || len(sh) != 2 {
		return nil, fmt.Errorf("l1.weight: %v", err2)
	}
	r.L1.Out, r.L1.In = sh[0], sh[1]
	r.H1Dim, r.InDim = sh[0], sh[1]
	if r.L1.B, _, err2 = readF32("l1.bias"); err2 != nil {
		return nil, err2
	}
	if r.L2.W, sh, err2 = readF32("l2.weight"); err2 != nil || len(sh) != 2 {
		return nil, fmt.Errorf("l2.weight: %v", err2)
	}
	r.L2.Out, r.L2.In = sh[0], sh[1]
	r.H2Dim = sh[0]
	if r.L2.B, _, err2 = readF32("l2.bias"); err2 != nil {
		return nil, err2
	}
	if r.L3.W, sh, err2 = readF32("l3.weight"); err2 != nil || len(sh) != 2 {
		return nil, fmt.Errorf("l3.weight: %v", err2)
	}
	r.L3.Out, r.L3.In = sh[0], sh[1]
	if r.L3.B, _, err2 = readF32("l3.bias"); err2 != nil {
		return nil, err2
	}
	if r.Skip.W, sh, err2 = readF32("skip.weight"); err2 != nil || len(sh) != 2 {
		return nil, fmt.Errorf("skip.weight: %v", err2)
	}
	r.Skip.Out, r.Skip.In = sh[0], sh[1]
	if r.Skip.B, _, err2 = readF32("skip.bias"); err2 != nil {
		return nil, err2
	}
	if rawMeta, ok := hdr["__metadata__"]; ok {
		var meta map[string]string
		if json.Unmarshal(rawMeta, &meta) == nil {
			r.Trained, _ = strconv.Atoi(meta["trained"])
			r.Step, _ = strconv.Atoi(meta["step"])
		}
	}
	r.L1.ensureOpt()
	r.L2.ensureOpt()
	r.L3.ensureOpt()
	r.Skip.ensureOpt()
	return r, nil
}
