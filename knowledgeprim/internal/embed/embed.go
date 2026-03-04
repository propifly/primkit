// Package embed provides an abstraction over embedding providers (Gemini,
// OpenAI, custom). The Embedder interface lets knowledgeprim generate vector
// embeddings for entities and search queries without coupling to a provider.
package embed

import (
	"context"
	"fmt"
)

// Embedder generates vector embeddings from text.
type Embedder interface {
	// Embed returns a vector embedding for the given text.
	Embed(ctx context.Context, text string) ([]float32, error)

	// Dimensions returns the expected output dimension count.
	Dimensions() int
}

// Config holds the embedding provider configuration.
type Config struct {
	Provider   string `yaml:"provider"`   // gemini, openai, custom
	Model      string `yaml:"model"`      // provider-specific model name
	Dimensions int    `yaml:"dimensions"` // expected output dimensions
	APIKey     string `yaml:"api_key"`    // API key
	Endpoint   string `yaml:"endpoint"`   // custom endpoint URL
}

// New creates an Embedder from config. Returns nil if no provider is configured.
func New(cfg Config) (Embedder, error) {
	if cfg.Provider == "" {
		return nil, nil
	}

	switch cfg.Provider {
	case "gemini":
		return NewGemini(cfg)
	case "openai":
		return NewOpenAI(cfg)
	case "custom":
		if cfg.Endpoint == "" {
			return nil, fmt.Errorf("custom embedding provider requires endpoint")
		}
		return NewOpenAI(cfg) // Custom uses OpenAI-compatible API
	default:
		return nil, fmt.Errorf("unknown embedding provider: %q", cfg.Provider)
	}
}
