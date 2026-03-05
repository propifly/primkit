package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Regression tests
// ---------------------------------------------------------------------------

// TestDBPathFromConfig verifies that storage.db in the YAML config file is
// used as the database path when --db and KNOWLEDGEPRIM_DB are both unset.
//
// Regression for: config was loaded after the DB path fallback chain, so
// cfg.Storage.DB was never consulted and the hardcoded default always won.
func TestDBPathFromConfig(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "knowledgeprim-test.db")
	configPath := filepath.Join(dir, "config.yaml")

	content := []byte("storage:\n  db: " + dbPath + "\n")
	require.NoError(t, os.WriteFile(configPath, content, 0644))

	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configPath, "stats"})
	err := root.Execute()
	require.NoError(t, err, "command should succeed using DB path from config")

	_, statErr := os.Stat(dbPath)
	assert.NoError(t, statErr, "DB file should be created at path from config, not the default")
}

// TestLoadEmbedConfig_EnvVarInterpolation verifies that ${VAR} references in
// embedding.api_key (and other embedding fields) are expanded before YAML
// parsing, consistent with how the shared config handles all other fields.
//
// Regression for: loadEmbedConfig called yaml.Unmarshal on raw bytes without
// calling config.InterpolateEnvVars first, so ${API_KEY} was passed literally
// to the embedding provider, causing 401 errors at runtime.
func TestLoadEmbedConfig_EnvVarInterpolation(t *testing.T) {
	t.Setenv("TEST_EMBED_API_KEY", "sk-test-abc123")

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := []byte("embedding:\n  provider: openai\n  api_key: ${TEST_EMBED_API_KEY}\n  model: text-embedding-3-small\n")
	require.NoError(t, os.WriteFile(configPath, content, 0644))

	cfg := loadEmbedConfig(configPath)
	assert.Equal(t, "sk-test-abc123", cfg.APIKey,
		"${TEST_EMBED_API_KEY} must be expanded before YAML parsing")
	assert.Equal(t, "openai", cfg.Provider)
	assert.Equal(t, "text-embedding-3-small", cfg.Model)
}

// TestLoadEmbedConfig_MissingEnvVar verifies that a missing ${VAR} reference
// resolves to empty string (same behaviour as the shared config layer).
func TestLoadEmbedConfig_MissingEnvVar(t *testing.T) {
	os.Unsetenv("TEST_EMBED_API_KEY_MISSING")

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := []byte("embedding:\n  provider: gemini\n  api_key: ${TEST_EMBED_API_KEY_MISSING}\n")
	require.NoError(t, os.WriteFile(configPath, content, 0644))

	cfg := loadEmbedConfig(configPath)
	assert.Empty(t, cfg.APIKey,
		"missing env var in ${...} must resolve to empty string, not the literal placeholder")
}

// TestLoadEmbedConfig_EnvOverrideWins verifies that KNOWLEDGEPRIM_EMBEDDING_API_KEY
// env var takes precedence over the value in the config file.
func TestLoadEmbedConfig_EnvOverrideWins(t *testing.T) {
	t.Setenv("TEST_EMBED_FILE_KEY", "sk-from-file")
	t.Setenv("KNOWLEDGEPRIM_EMBEDDING_API_KEY", "sk-from-env")

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := []byte("embedding:\n  provider: openai\n  api_key: ${TEST_EMBED_FILE_KEY}\n")
	require.NoError(t, os.WriteFile(configPath, content, 0644))

	cfg := loadEmbedConfig(configPath)
	assert.Equal(t, "sk-from-env", cfg.APIKey,
		"KNOWLEDGEPRIM_EMBEDDING_API_KEY env var must override config file value")
}
