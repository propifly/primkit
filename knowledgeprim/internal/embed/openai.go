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

// OpenAIEmbedder generates embeddings using the OpenAI-compatible API.
// Also used for custom endpoints that implement the same interface.
type OpenAIEmbedder struct {
	apiKey     string
	model      string
	dimensions int
	endpoint   string
	client     *http.Client
}

// NewOpenAI creates an OpenAI (or compatible) embedding provider.
func NewOpenAI(cfg Config) (*OpenAIEmbedder, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("openai embedding provider requires api_key")
	}
	model := cfg.Model
	if model == "" {
		model = "text-embedding-3-small"
	}
	dims := cfg.Dimensions
	if dims == 0 {
		dims = 1536
	}
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/embeddings"
	}
	return &OpenAIEmbedder{
		apiKey:     cfg.APIKey,
		model:      model,
		dimensions: dims,
		endpoint:   endpoint,
		client:     &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (o *OpenAIEmbedder) Dimensions() int { return o.dimensions }

func (o *OpenAIEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	body := map[string]interface{}{
		"input": text,
		"model": o.model,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if len(result.Data) == 0 || len(result.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	return result.Data[0].Embedding, nil
}
