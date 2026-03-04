package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()
	assert.Equal(t, 8090, cfg.Server.Port, "default port should be 8090")
	assert.Empty(t, cfg.Storage.DB, "default db path should be empty (resolved later)")
	assert.False(t, cfg.Storage.Replicate.Enabled, "replication should be off by default")
}

func TestLoad_EmptyPath(t *testing.T) {
	cfg, err := Load("")
	require.NoError(t, err)
	assert.Equal(t, Defaults(), *cfg, "empty path should return defaults")
}

func TestLoad_ValidYAML(t *testing.T) {
	cfg, err := Load(filepath.Join("testdata", "valid.yaml"))
	require.NoError(t, err)

	assert.Equal(t, "/tmp/test.db", cfg.Storage.DB)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Len(t, cfg.Auth.Keys, 1)
	assert.Equal(t, "tp_sk_testkey123", cfg.Auth.Keys[0].Key)
	assert.Equal(t, "test-agent", cfg.Auth.Keys[0].Name)
}

func TestLoad_EnvVarInterpolation(t *testing.T) {
	// Set an env var that the test fixture references.
	t.Setenv("TEST_SECRET_KEY", "my-secret-value")

	cfg, err := Load(filepath.Join("testdata", "envvars.yaml"))
	require.NoError(t, err)

	assert.Equal(t, "my-secret-value", cfg.Storage.Replicate.SecretAccessKey,
		"${TEST_SECRET_KEY} should be replaced with env var value")
}

func TestLoad_MissingEnvVar(t *testing.T) {
	// Ensure the referenced var is NOT set.
	os.Unsetenv("TEST_SECRET_KEY")

	cfg, err := Load(filepath.Join("testdata", "envvars.yaml"))
	require.NoError(t, err)

	assert.Empty(t, cfg.Storage.Replicate.SecretAccessKey,
		"missing env var should resolve to empty string")
}

func TestLoad_NonexistentFile(t *testing.T) {
	_, err := Load("/tmp/does-not-exist-config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reading config file")
}

func TestLoad_MalformedYAML(t *testing.T) {
	// Write a temporary malformed YAML file.
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	require.NoError(t, os.WriteFile(path, []byte("{{invalid yaml"), 0644))

	_, err := Load(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parsing config file")
}

func TestLoadWithEnvOverrides(t *testing.T) {
	t.Setenv("TASKPRIM_DB", "/override/path.db")
	t.Setenv("TASKPRIM_SERVER_PORT", "7777")

	cfg, err := LoadWithEnvOverrides(filepath.Join("testdata", "valid.yaml"), "TASKPRIM")
	require.NoError(t, err)

	assert.Equal(t, "/override/path.db", cfg.Storage.DB,
		"TASKPRIM_DB should override storage.db from file")
	assert.Equal(t, 7777, cfg.Server.Port,
		"TASKPRIM_SERVER_PORT should override server.port from file")
}

func TestLoadWithEnvOverrides_ReplicateEnabled(t *testing.T) {
	t.Setenv("TASKPRIM_REPLICATE_ENABLED", "true")
	t.Setenv("TASKPRIM_REPLICATE_PROVIDER", "r2")
	t.Setenv("TASKPRIM_REPLICATE_BUCKET", "my-bucket")

	cfg, err := LoadWithEnvOverrides("", "TASKPRIM")
	require.NoError(t, err)

	assert.True(t, cfg.Storage.Replicate.Enabled)
	assert.Equal(t, "r2", cfg.Storage.Replicate.Provider)
	assert.Equal(t, "my-bucket", cfg.Storage.Replicate.Bucket)
}

func TestLoadWithEnvOverrides_InvalidPort(t *testing.T) {
	t.Setenv("TASKPRIM_SERVER_PORT", "not-a-number")

	cfg, err := LoadWithEnvOverrides("", "TASKPRIM")
	require.NoError(t, err)

	// Invalid port should be silently ignored, keeping the default.
	assert.Equal(t, 8090, cfg.Server.Port)
}

func TestInterpolateEnvVars(t *testing.T) {
	t.Setenv("FOO", "bar")
	t.Setenv("BAZ", "qux")

	input := []byte("key1: ${FOO}\nkey2: ${BAZ}\nkey3: ${MISSING}")
	result := interpolateEnvVars(input)

	assert.Contains(t, string(result), "key1: bar")
	assert.Contains(t, string(result), "key2: qux")
	assert.Contains(t, string(result), "key3: ")
}
