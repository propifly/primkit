// Package config loads YAML configuration with environment variable interpolation
// and override support. Both taskprim and stateprim use this package.
//
// Configuration is resolved in three layers (lowest to highest precedence):
//  1. Hardcoded defaults in code
//  2. Values from config.yaml (with ${ENV_VAR} interpolation)
//  3. Environment variable overrides (TASKPRIM_* or STATEPRIM_*)
package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration shared by all primitives. Each primitive
// may extend this with its own section (e.g., taskprim.default_list).
type Config struct {
	Storage StorageConfig `yaml:"storage"`
	Auth    AuthConfig    `yaml:"auth"`
	Server  ServerConfig  `yaml:"server"`
}

// StorageConfig controls where and how the SQLite database is stored.
type StorageConfig struct {
	DB        string          `yaml:"db"`
	Replicate ReplicateConfig `yaml:"replicate"`
}

// ReplicateConfig controls Litestream replication to object storage.
// When enabled, WAL frames are continuously streamed to the target.
type ReplicateConfig struct {
	Enabled         bool   `yaml:"enabled"`
	Provider        string `yaml:"provider"` // r2, s3, b2, gcs
	Bucket          string `yaml:"bucket"`
	Path            string `yaml:"path"`
	Endpoint        string `yaml:"endpoint"` // required for R2
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
}

// AuthConfig holds API keys for serve and MCP modes.
type AuthConfig struct {
	Keys []KeyConfig `yaml:"keys"`
}

// KeyConfig maps an API key to a human-readable name. The name is used as the
// task source when tasks are created via the API (so you know "johanna created
// this" vs "ios-app created this").
type KeyConfig struct {
	Key  string `yaml:"key"`
	Name string `yaml:"name"`
}

// ServerConfig controls the HTTP server in serve and MCP SSE modes.
type ServerConfig struct {
	Port int `yaml:"port"`
}

// Defaults returns a Config with sensible defaults for local development.
func Defaults() Config {
	return Config{
		Server: ServerConfig{Port: 8090},
	}
}

// envVarPattern matches ${VAR_NAME} references in YAML values.
var envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// Load reads a YAML config file, interpolates ${ENV_VAR} references with their
// values from the environment, and returns the parsed config. If path is empty,
// returns defaults.
func Load(path string) (*Config, error) {
	cfg := Defaults()

	if path == "" {
		return &cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Replace ${ENV_VAR} references before parsing YAML. This allows secrets
	// to live in the environment rather than in the config file.
	data = interpolateEnvVars(data)

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &cfg, nil
}

// LoadWithEnvOverrides loads a config file and optionally applies environment
// variable overrides based on whether a config file was explicitly provided.
//
// Behaviour depends on whether path is set:
//
//   - No config file (path == ""): env vars with the given prefix are applied as
//     the primary configuration mechanism. Useful for container and CI deployments
//     where mounting a config file is impractical.
//     Effective precedence: env var → hardcoded default.
//
//   - Config file provided (path != ""): the file is authoritative and env var
//     overrides are NOT applied. This is the correct behaviour for multi-agent
//     deployments where each agent passes its own --config file — a global env
//     var (e.g. TASKPRIM_DB=/shared.db) must not silently override per-agent
//     configuration. Use ${VAR} interpolation inside the YAML to inject secrets
//     or other dynamic values while still keeping the file authoritative.
//     Effective precedence: config file → hardcoded default.
//
// In both cases the --db CLI flag takes top priority (handled in root.go before
// this function is called).
func LoadWithEnvOverrides(path, prefix string) (*Config, error) {
	cfg, err := Load(path)
	if err != nil {
		return nil, err
	}

	// Env overrides are the no-config fallback, not a global override on top of
	// an explicit config file. When a config file is present it wins.
	if path == "" {
		applyEnvOverrides(cfg, prefix)
	}

	return cfg, nil
}

// InterpolateEnvVars replaces all ${VAR_NAME} patterns in the raw YAML bytes
// with corresponding environment variable values. Missing vars resolve to "".
// Exported so callers that parse YAML outside the standard Load path (e.g.
// extended config sections) can apply the same expansion.
func InterpolateEnvVars(data []byte) []byte {
	return envVarPattern.ReplaceAllFunc(data, func(match []byte) []byte {
		// Extract variable name from ${VAR_NAME}.
		varName := string(match[2 : len(match)-1])
		return []byte(os.Getenv(varName))
	})
}

// interpolateEnvVars is the unexported alias kept for internal use.
func interpolateEnvVars(data []byte) []byte { return InterpolateEnvVars(data) }

// applyEnvOverrides checks for environment variables with the given prefix and
// fills in config fields. Called only when no config file was provided (see
// LoadWithEnvOverrides). For example, with prefix "TASKPRIM":
//   - TASKPRIM_DB          → storage.db
//   - TASKPRIM_SERVER_PORT → server.port
func applyEnvOverrides(cfg *Config, prefix string) {
	if v := os.Getenv(prefix + "_DB"); v != "" {
		cfg.Storage.DB = v
	}
	if v := os.Getenv(prefix + "_SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv(prefix + "_REPLICATE_ENABLED"); v != "" {
		cfg.Storage.Replicate.Enabled = strings.EqualFold(v, "true")
	}
	if v := os.Getenv(prefix + "_REPLICATE_PROVIDER"); v != "" {
		cfg.Storage.Replicate.Provider = v
	}
	if v := os.Getenv(prefix + "_REPLICATE_BUCKET"); v != "" {
		cfg.Storage.Replicate.Bucket = v
	}
	if v := os.Getenv(prefix + "_REPLICATE_PATH"); v != "" {
		cfg.Storage.Replicate.Path = v
	}
	if v := os.Getenv(prefix + "_REPLICATE_ENDPOINT"); v != "" {
		cfg.Storage.Replicate.Endpoint = v
	}
	if v := os.Getenv(prefix + "_REPLICATE_ACCESS_KEY_ID"); v != "" {
		cfg.Storage.Replicate.AccessKeyID = v
	}
	if v := os.Getenv(prefix + "_REPLICATE_SECRET_ACCESS_KEY"); v != "" {
		cfg.Storage.Replicate.SecretAccessKey = v
	}
}
