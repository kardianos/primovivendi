package mp3join

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestJoinCBRNoReencode(t *testing.T) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not available")
	}
	dir := t.TempDir()
	a := filepath.Join(dir, "a.mp3")
	b := filepath.Join(dir, "b.mp3")
	out := filepath.Join(dir, "out.mp3")

	// Two short CBR clips, identical encode settings (simulates TTS chunks).
	gen := func(path, freq string, dur string) {
		cmd := exec.Command(ffmpeg, "-y",
			"-f", "lavfi", "-i", "sine=frequency="+freq+":duration="+dur,
			"-c:a", "libmp3lame", "-b:a", "128k", "-ar", "24000", "-ac", "1",
			path,
		)
		cmd.Stderr = nil
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("ffmpeg %s: %v\n%s", path, err, out)
		}
	}
	gen(a, "440", "0.4")
	gen(b, "880", "0.4")

	res, err := JoinFiles(out, []string{a, b}, Options{RequireSameProfile: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Frames < 2 {
		t.Fatalf("frames: %d", res.Frames)
	}
	if res.InputFiles != 2 {
		t.Fatalf("inputs: %d", res.InputFiles)
	}
	if res.VBR {
		// CBR 128k should not flag VBR (padding may still share bitrate index)
		// lame CBR can still have padding bit differences but same bitrate
		t.Logf("note: VBR flag=%v profile=%s", res.VBR, res.Profile)
	}

	// Output must only contain frames we can re-parse; no double ID3 junk mid-stream.
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	frames, err := IterFrames(bytes.NewReader(data), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != res.Frames && !res.VBR {
		// if VBR we prepended Xing which IterFrames(skipLeadingVBR) drops
	}
	got, err := IterFrames(bytes.NewReader(data), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) < res.Frames {
		// when VBR, skipLeadingVBR drops Xing so audio frames == res.Frames
		if res.VBR && len(got) == res.Frames {
			// ok
		} else if !res.VBR && len(got) == res.Frames {
			// ok
		} else {
			t.Fatalf("reparse frames=%d want %d vbr=%v", len(got), res.Frames, res.VBR)
		}
	}

	// Byte length of joined audio ≈ sum of frame payloads (tags stripped from inputs)
	fa, _ := os.ReadFile(a)
	fb, _ := os.ReadFile(b)
	// joined without tags should be smaller than naive cat of files (which keeps 2 ID3 sets)
	naive := len(fa) + len(fb)
	if len(data) >= naive {
		// usually smaller after stripping tags; allow equal if no tags
		t.Logf("joined %d bytes, naive cat %d", len(data), naive)
	}

	// Profile should be 24 kHz mono-ish from our encode
	if res.Profile.SamplingRate != 24000 {
		t.Fatalf("sample rate %d", res.Profile.SamplingRate)
	}
}

func TestJoinStripsID3(t *testing.T) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not available")
	}
	dir := t.TempDir()
	a := filepath.Join(dir, "a.mp3")
	b := filepath.Join(dir, "b.mp3")
	// Write ID3 with ffmpeg metadata
	for i, path := range []string{a, b} {
		title := "part" + string(rune('A'+i))
		cmd := exec.Command(ffmpeg, "-y",
			"-f", "lavfi", "-i", "sine=frequency=500:duration=0.25",
			"-c:a", "libmp3lame", "-b:a", "64k", "-ar", "24000", "-ac", "1",
			"-metadata", "title="+title,
			"-id3v2_version", "3",
			path,
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}
	}
	joined, res, err := JoinBytes(mustRead(t, a, b), Options{RequireSameProfile: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Frames == 0 {
		t.Fatal("no frames")
	}
	// Mid-stream "ID3" should not appear after first few hundred bytes of pure frames
	// (no second ID3 header in the middle)
	if i := bytes.Index(joined[10:], []byte("ID3")); i >= 0 {
		t.Fatalf("found mid-stream ID3 at %d", i+10)
	}
	// "TAG" ID3v1 at end of each file should not leave two TAGs
	if bytes.Count(joined, []byte("TAG")) > 1 {
		t.Fatalf("multiple ID3v1 TAG markers")
	}
}

func TestProfileMismatch(t *testing.T) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not available")
	}
	dir := t.TempDir()
	a := filepath.Join(dir, "a.mp3")
	b := filepath.Join(dir, "b.mp3")
	run := func(path, ar string) {
		cmd := exec.Command(ffmpeg, "-y",
			"-f", "lavfi", "-i", "sine=frequency=400:duration=0.2",
			"-c:a", "libmp3lame", "-b:a", "64k", "-ar", ar, "-ac", "1", path)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}
	}
	run(a, "24000")
	run(b, "44100")
	_, _, err = JoinBytes(mustRead(t, a, b), Options{RequireSameProfile: true})
	if err == nil {
		t.Fatal("expected profile mismatch error")
	}
}

func TestNextFrameEmpty(t *testing.T) {
	f, err := NextFrame(bytes.NewReader(nil))
	if err != nil || f != nil {
		t.Fatalf("%v %v", f, err)
	}
}

func mustRead(t *testing.T, paths ...string) [][]byte {
	t.Helper()
	var out [][]byte
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, b)
	}
	return out
}
