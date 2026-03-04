package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GeminiEmbedder generates embeddings using Google's Gemini API.
type GeminiEmbedder struct {
	apiKey     string
	model      string
	dimensions int
	client     *http.Client
}

// NewGemini creates a Gemini embedding provider.
func NewGemini(cfg Config) (*GeminiEmbedder, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("gemini embedding provider requires api_key")
	}
	model := cfg.Model
	if model == "" {
		model = "text-embedding-004"
	}
	dims := cfg.Dimensions
	if dims == 0 {
		dims = 768
	}
	return &GeminiEmbedder{
		apiKey:     cfg.APIKey,
		model:      model,
		dimensions: dims,
		client:     &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (g *GeminiEmbedder) Dimensions() int { return g.dimensions }
func (g *GeminiEmbedder) Provider() string  { return "gemini" }
func (g *GeminiEmbedder) Model() string     { return g.model }

func (g *GeminiEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s",
		g.model, g.apiKey,
	)

	body := map[string]interface{}{
		"model": fmt.Sprintf("models/%s", g.model),
		"content": map[string]interface{}{
			"parts": []map[string]string{
				{"text": text},
			},
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Embedding struct {
			Values []float32 `json:"values"`
		} `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if len(result.Embedding.Values) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	return result.Embedding.Values, nil
}
