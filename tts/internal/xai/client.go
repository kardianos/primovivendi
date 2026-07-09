package xai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultEndpoint is the unary TTS REST path.
const DefaultEndpoint = "https://api.x.ai/v1/tts"

// DefaultVoice is the voice_id used when none is specified.
// Carina is the project default for book narration (override with -voice).
const DefaultVoice = "carina"

// DefaultLanguage is BCP-47 language for synthesis (English for this book).
const DefaultLanguage = "en"

const defaultTimeout = 15 * time.Minute

// DefaultMP3 is a stable output profile so multi-chunk joins stay same-settings.
var DefaultMP3 = &OutputFormat{
	Codec:      "mp3",
	SampleRate: 24000,
	BitRate:    128000,
}

// Request is the unary TTS JSON body (subset used by this tool).
type Request struct {
	Text              string        `json:"text"`
	VoiceID           string        `json:"voice_id,omitempty"`
	Language          string        `json:"language"`
	Speed             float64       `json:"speed,omitempty"`
	TextNormalization bool          `json:"text_normalization,omitempty"`
	OutputFormat      *OutputFormat `json:"output_format,omitempty"`
}

// OutputFormat selects codec parameters. Keep identical across chunks if
// you plan to concatenate MP3 frames via mp3join.
type OutputFormat struct {
	Codec      string `json:"codec,omitempty"`
	SampleRate int    `json:"sample_rate,omitempty"`
	BitRate    int    `json:"bit_rate,omitempty"`
}

// Client talks to the xAI TTS REST API.
type Client struct {
	APIKey     string
	Endpoint   string
	HTTPClient *http.Client
}

// NewClientFromEnv is an alias for NewClient (env or .ttskey).
func NewClientFromEnv() (*Client, error) {
	return NewClient()
}

// Synthesize sends text to TTS and returns raw audio bytes (e.g. MP3).
func (c *Client) Synthesize(ctx context.Context, req Request) ([]byte, error) {
	if c == nil {
		return nil, fmt.Errorf("nil client")
	}
	if c.APIKey == "" {
		return nil, fmt.Errorf("missing API key")
	}
	if req.Language == "" {
		req.Language = DefaultLanguage
	}
	if req.VoiceID == "" {
		req.VoiceID = DefaultVoice
	}
	if req.OutputFormat == nil {
		req.OutputFormat = DefaultMP3
	}
	endpoint := c.Endpoint
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	hc := c.HTTPClient
	if hc == nil {
		hc = httpClient()
	}
	resp, err := hc.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tts %s: %s", resp.Status, truncate(string(data), 500))
	}
	return data, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
