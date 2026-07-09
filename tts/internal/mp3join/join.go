package mp3join

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// Options controls join behavior.
type Options struct {
	// RequireSameProfile rejects inputs whose first audio frame differs in
	// sample rate / channel mode / MPEG version / layer from the first file.
	// Bitrate may still vary (VBR); when false, any stream is accepted.
	RequireSameProfile bool
}

// Result is metadata about a completed join.
type Result struct {
	Frames     int
	Bytes      int
	VBR        bool
	Profile    Profile
	InputFiles int
}

// JoinReaders concatenates MP3 streams from readers into w without re-encoding.
func JoinReaders(w io.Writer, readers []io.Reader, opt Options) (Result, error) {
	var res Result
	if len(readers) == 0 {
		return res, fmt.Errorf("mp3join: no inputs")
	}

	var (
		buf          bytes.Buffer
		firstBitRate int
		haveProfile  bool
	)

	for i, r := range readers {
		frames, err := IterFrames(r, true)
		if err != nil {
			return res, fmt.Errorf("mp3join: input %d: %w", i, err)
		}
		if len(frames) == 0 {
			return res, fmt.Errorf("mp3join: input %d: no MP3 frames", i)
		}
		res.InputFiles++

		for _, f := range frames {
			p := ProfileOf(f)
			if !haveProfile {
				res.Profile = p
				haveProfile = true
				firstBitRate = f.BitRate
			} else {
				if opt.RequireSameProfile {
					if p.SamplingRate != res.Profile.SamplingRate ||
						p.ChannelMode != res.Profile.ChannelMode ||
						p.MPEGVersion != res.Profile.MPEGVersion ||
						p.MPEGLayer != res.Profile.MPEGLayer {
						return res, fmt.Errorf("mp3join: input %d profile %s != %s",
							i, p, res.Profile)
					}
				}
				if f.BitRate != firstBitRate {
					res.VBR = true
				}
			}
			if _, err := buf.Write(f.Raw); err != nil {
				return res, err
			}
			res.Frames++
			res.Bytes += len(f.Raw)
		}
	}

	if res.VBR {
		xh := newXingHeader(uint32(res.Frames), uint32(res.Bytes))
		if _, err := w.Write(xh.Raw); err != nil {
			return res, err
		}
	}
	_, err := io.Copy(w, &buf)
	return res, err
}

// JoinFiles concatenates on-disk MP3 files into outPath.
func JoinFiles(outPath string, inPaths []string, opt Options) (Result, error) {
	if len(inPaths) == 0 {
		return Result{}, fmt.Errorf("mp3join: no input paths")
	}
	readers := make([]io.Reader, 0, len(inPaths))
	files := make([]*os.File, 0, len(inPaths))
	defer func() {
		for _, f := range files {
			_ = f.Close()
		}
	}()
	for _, p := range inPaths {
		f, err := os.Open(p)
		if err != nil {
			return Result{}, err
		}
		files = append(files, f)
		readers = append(readers, f)
	}

	out, err := os.Create(outPath)
	if err != nil {
		return Result{}, err
	}
	defer out.Close()

	res, err := JoinReaders(out, readers, opt)
	if err != nil {
		_ = os.Remove(outPath)
		return res, err
	}
	return res, out.Close()
}

// JoinBytes concatenates in-memory MP3 blobs.
func JoinBytes(parts [][]byte, opt Options) ([]byte, Result, error) {
	readers := make([]io.Reader, len(parts))
	for i, p := range parts {
		readers[i] = bytes.NewReader(p)
	}
	var buf bytes.Buffer
	res, err := JoinReaders(&buf, readers, opt)
	return buf.Bytes(), res, err
}
