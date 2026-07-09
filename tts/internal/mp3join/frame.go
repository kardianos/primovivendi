// Package mp3join concatenates MP3 bitstreams without re-encoding.
//
// Approach (same idea as dmulholl/mp3lib / mp3cat):
//  1. Scan for objects: ID3v2 tags, ID3v1 tags, or MPEG frame headers.
//  2. Drop all tags and any leading Xing/Info/VBRI (VBR) header frames.
//  3. Append raw audio frame bytes from each input in order.
//  4. If multiple bitrates appear, prepend a simple Xing header so duration
//     estimators work; CBR streams from a fixed TTS profile need none.
//
// Frame length uses the standard formula from the MPEG header fields.
// Garbage between objects is skipped one byte at a time until resync.
package mp3join

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Frame is one MPEG audio frame with its full on-wire bytes (header + body).
type Frame struct {
	MPEGVersion  byte
	MPEGLayer    byte
	BitRate      int // bits per second
	SamplingRate int // Hz
	ChannelMode  byte
	SampleCount  int
	FrameLength  int
	Raw          []byte
}

// MPEG version / layer / channel constants (header bit encodings).
const (
	mpegVersion2_5 = 0
	mpegVersionRes = 1
	mpegVersion2   = 2
	mpegVersion1   = 3

	mpegLayerRes = 0
	mpegLayerIII = 1
	mpegLayerII  = 2
	mpegLayerI   = 3

	chStereo      = 0
	chJointStereo = 1
	chDual        = 2
	chMono        = 3
)

var (
	v1l1BR = []int{0, 32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 416, 448}
	v1l2BR = []int{0, 32, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 384}
	v1l3BR = []int{0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320}
	v2l1BR = []int{0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256}
	v2l23  = []int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160}

	v1SR  = []int{44100, 48000, 32000}
	v2SR  = []int{22050, 24000, 16000}
	v25SR = []int{11025, 12000, 8000}
)

// NextFrame returns the next audio frame, skipping ID3 tags and junk.
// Returns (nil, nil) at EOF.
func NextFrame(r io.Reader) (*Frame, error) {
	for {
		obj, err := nextObject(r)
		if err != nil {
			return nil, err
		}
		if obj == nil {
			return nil, nil
		}
		if f, ok := obj.(*Frame); ok {
			return f, nil
		}
		// ID3 tags discarded
	}
}

type id3v1 struct{ raw []byte }
type id3v2 struct{ raw []byte }

func nextObject(r io.Reader) (any, error) {
	hdr := make([]byte, 4)
	if err := readFull(r, hdr); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, nil
		}
		return nil, err
	}
	last := hdr[3:]

	for {
		// ID3v1: "TAG" + 125 more bytes = 128 total
		if hdr[0] == 'T' && hdr[1] == 'A' && hdr[2] == 'G' {
			raw := make([]byte, 128)
			copy(raw, hdr)
			if err := readFull(r, raw[4:]); err != nil {
				return nil, err
			}
			return &id3v1{raw: raw}, nil
		}

		// ID3v2: "ID3" + 7 more header bytes; size is syncsafe int
		if hdr[0] == 'I' && hdr[1] == 'D' && hdr[2] == '3' {
			rest := make([]byte, 6)
			if err := readFull(r, rest); err != nil {
				return nil, err
			}
			size := int(rest[2])<<21 | int(rest[3])<<14 | int(rest[4])<<7 | int(rest[5])
			if size < 0 || size > 50<<20 { // sanity: 50 MiB tag max
				// treat as junk, resync
			} else {
				raw := make([]byte, 10+size)
				copy(raw, hdr)
				copy(raw[4:], rest)
				if err := readFull(r, raw[10:]); err != nil {
					return nil, err
				}
				return &id3v2{raw: raw}, nil
			}
		}

		// MPEG frame sync: 11 set bits
		if hdr[0] == 0xFF && hdr[1]&0xE0 == 0xE0 {
			f := &Frame{}
			if parseHeader(hdr, f) {
				f.Raw = make([]byte, f.FrameLength)
				copy(f.Raw, hdr)
				if err := readFull(r, f.Raw[4:]); err != nil {
					return nil, err
				}
				return f, nil
			}
		}

		// Resync: shift one byte
		hdr[0], hdr[1], hdr[2] = hdr[1], hdr[2], hdr[3]
		n, err := r.Read(last)
		if n < 1 {
			if err == io.EOF {
				return nil, nil
			}
			if err != nil {
				return nil, err
			}
			return nil, nil
		}
	}
}

func parseHeader(h []byte, f *Frame) bool {
	f.MPEGVersion = (h[1] & 0x18) >> 3
	if f.MPEGVersion == mpegVersionRes {
		return false
	}
	f.MPEGLayer = (h[1] & 0x06) >> 1
	if f.MPEGLayer == mpegLayerRes {
		return false
	}

	brIdx := int((h[2] & 0xF0) >> 4)
	if brIdx == 0 || brIdx == 15 {
		return false
	}
	srIdx := int((h[2] & 0x0C) >> 2)
	if srIdx == 3 {
		return false
	}

	if f.MPEGVersion == mpegVersion1 {
		switch f.MPEGLayer {
		case mpegLayerI:
			f.BitRate = v1l1BR[brIdx] * 1000
		case mpegLayerII:
			f.BitRate = v1l2BR[brIdx] * 1000
		case mpegLayerIII:
			f.BitRate = v1l3BR[brIdx] * 1000
		}
		f.SamplingRate = v1SR[srIdx]
	} else {
		switch f.MPEGLayer {
		case mpegLayerI:
			f.BitRate = v2l1BR[brIdx] * 1000
		case mpegLayerII, mpegLayerIII:
			f.BitRate = v2l23[brIdx] * 1000
		}
		if f.MPEGVersion == mpegVersion2 {
			f.SamplingRate = v2SR[srIdx]
		} else {
			f.SamplingRate = v25SR[srIdx]
		}
	}

	padding := 0
	if h[2]&0x02 != 0 {
		if f.MPEGLayer == mpegLayerI {
			padding = 4
		} else {
			padding = 1
		}
	}

	f.ChannelMode = (h[3] & 0xC0) >> 6
	modeExt := (h[3] & 0x30) >> 4
	if f.ChannelMode != chJointStereo && modeExt != 0 {
		return false
	}
	if h[3]&0x03 == 2 { // reserved emphasis
		return false
	}

	// samples per frame
	if f.MPEGVersion == mpegVersion1 {
		switch f.MPEGLayer {
		case mpegLayerI:
			f.SampleCount = 384
		default:
			f.SampleCount = 1152
		}
	} else {
		switch f.MPEGLayer {
		case mpegLayerI:
			f.SampleCount = 384
		case mpegLayerII:
			f.SampleCount = 1152
		case mpegLayerIII:
			f.SampleCount = 576
		}
	}

	// frame_length includes the 4-byte header
	f.FrameLength = (f.SampleCount/8)*f.BitRate/f.SamplingRate + padding
	if f.FrameLength < 4 {
		return false
	}
	return true
}

func sideInfoSize(f *Frame) int {
	if f.MPEGLayer != mpegLayerIII {
		return 0
	}
	if f.MPEGVersion == mpegVersion1 {
		if f.ChannelMode == chMono {
			return 17
		}
		return 32
	}
	if f.ChannelMode == chMono {
		return 9
	}
	return 17
}

// IsVBRHeader reports Xing/Info or VBRI metadata frames (not real audio).
func IsVBRHeader(f *Frame) bool {
	if f == nil {
		return false
	}
	side := sideInfoSize(f)
	// Xing / Info after side info
	off := 4 + side
	if len(f.Raw) >= off+4 {
		id := f.Raw[off : off+4]
		if bytes.Equal(id, []byte("Xing")) || bytes.Equal(id, []byte("Info")) {
			return true
		}
	}
	// VBRI at fixed offset 32 after header
	if len(f.Raw) >= 4+32+4 {
		if bytes.Equal(f.Raw[4+32:4+32+4], []byte("VBRI")) {
			return true
		}
	}
	return false
}

// IterFrames yields every audio frame from r, skipping tags and optional
// first-frame VBR headers when skipLeadingVBR is true.
func IterFrames(r io.Reader, skipLeadingVBR bool) ([]*Frame, error) {
	var out []*Frame
	first := true
	for {
		f, err := NextFrame(r)
		if err != nil {
			return nil, err
		}
		if f == nil {
			return out, nil
		}
		if first {
			first = false
			if skipLeadingVBR && IsVBRHeader(f) {
				continue
			}
		}
		out = append(out, f)
	}
}

func readFull(r io.Reader, buf []byte) error {
	_, err := io.ReadFull(r, buf)
	return err
}

// newXingHeader builds a minimal Xing frame (MPEG1 Layer III stereo template)
// carrying total frame and byte counts for VBR duration.
func newXingHeader(totalFrames, totalBytes uint32) *Frame {
	// Template: MPEG1 Layer3, 128 kbps, 44100 Hz, stereo — common enough for seeking metadata.
	// Actual audio frames may differ; Xing is only advisory for VBR files.
	raw := make([]byte, 417) // 1152/8 * 128000/44100 = 417.959 → 417 without padding
	raw[0] = 0xFF
	raw[1] = 0xFB // version1 layer3 no CRC
	raw[2] = 0x90 // 128kbps, 44100, no pad
	raw[3] = 0xC0 // stereo
	f := &Frame{}
	if !parseHeader(raw[:4], f) {
		// fall back length
		f.FrameLength = 417
		f.MPEGLayer = mpegLayerIII
		f.MPEGVersion = mpegVersion1
		f.ChannelMode = chStereo
		f.BitRate = 128000
		f.SamplingRate = 44100
		f.SampleCount = 1152
	}
	f.Raw = make([]byte, f.FrameLength)
	copy(f.Raw, raw[:4])

	off := 4 + sideInfoSize(f)
	if off+16 > len(f.Raw) {
		// ensure room
		return f
	}
	copy(f.Raw[off:off+4], []byte("Xing"))
	f.Raw[off+7] = 3 // frames + bytes flags
	binary.BigEndian.PutUint32(f.Raw[off+8:off+12], totalFrames)
	binary.BigEndian.PutUint32(f.Raw[off+12:off+16], totalBytes)
	return f
}

// Profile describes stream parameters used for same-settings checks.
type Profile struct {
	BitRate      int
	SamplingRate int
	ChannelMode  byte
	MPEGVersion  byte
	MPEGLayer    byte
}

func (p Profile) String() string {
	return fmt.Sprintf("br=%d sr=%d ch=%d v=%d L=%d",
		p.BitRate, p.SamplingRate, p.ChannelMode, p.MPEGVersion, p.MPEGLayer)
}

// ProfileOf returns the encoding profile of a frame.
func ProfileOf(f *Frame) Profile {
	return Profile{
		BitRate:      f.BitRate,
		SamplingRate: f.SamplingRate,
		ChannelMode:  f.ChannelMode,
		MPEGVersion:  f.MPEGVersion,
		MPEGLayer:    f.MPEGLayer,
	}
}
