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
	require.NoError(t, os.WriteFile(path, []byte("{{invalid yaml"), 0o644))

	_, err := Load(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parsing config file")
}

// TestLoadWithEnvOverrides_ConfigFileWinsWhenPathSet verifies that when a config
// file path is provided, env var overrides are NOT applied — the config file is
// authoritative. This is the correct behaviour for multi-agent deployments where
// each agent has its own --config file.
func TestLoadWithEnvOverrides_ConfigFileWinsWhenPathSet(t *testing.T) {
	t.Setenv("TASKPRIM_DB", "/override/path.db")
	t.Setenv("TASKPRIM_SERVER_PORT", "7777")

	cfg, err := LoadWithEnvOverrides(filepath.Join("testdata", "valid.yaml"), "TASKPRIM")
	require.NoError(t, err)

	// Config file must win — env vars are not applied when a file is present.
	assert.Equal(t, "/tmp/test.db", cfg.Storage.DB,
		"config file storage.db must win over TASKPRIM_DB when --config is provided")
	assert.Equal(t, 9090, cfg.Server.Port,
		"config file server.port must win over TASKPRIM_SERVER_PORT when --config is provided")
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
	result := InterpolateEnvVars(input)

	assert.Contains(t, string(result), "key1: bar")
	assert.Contains(t, string(result), "key2: qux")
	assert.Contains(t, string(result), "key3: ")
}

// Regression test: storage.db in YAML config must be loaded and accessible.
// Bug: stateprim/taskprim/knowledgeprim resolved DB path before loading config,
// so cfg.Storage.DB was never consulted. Verify the config layer reads it.
func TestLoad_StorageDBFromConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("storage:\n  db: /custom/path/to.db\n")
	require.NoError(t, os.WriteFile(path, content, 0o644))

	cfg, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, "/custom/path/to.db", cfg.Storage.DB,
		"storage.db from YAML must be loaded into cfg.Storage.DB")
}

// Regression test: ALL R2 fields (bucket, endpoint, access_key_id, secret_access_key)
// must expand from ${...} references in the YAML config. Previously only
// secret_access_key was covered by the test fixture; bucket and endpoint were
// left as empty strings in the knowledgeprim example config with no ${...} reference,
// making them impossible to set via env var interpolation.
func TestLoad_R2AllFieldsInterpolated(t *testing.T) {
	t.Setenv("R2_BUCKET", "my-test-bucket")
	t.Setenv("R2_ENDPOINT", "https://abc123.r2.cloudflarestorage.com")
	t.Setenv("R2_ACCESS_KEY_ID", "test-access-key")
	t.Setenv("R2_SECRET_ACCESS_KEY", "test-secret-key")

	cfg, err := Load(filepath.Join("testdata", "r2vars.yaml"))
	require.NoError(t, err)

	assert.Equal(t, "my-test-bucket", cfg.Storage.Replicate.Bucket,
		"${R2_BUCKET} must be expanded in storage.replicate.bucket")
	assert.Equal(t, "https://abc123.r2.cloudflarestorage.com", cfg.Storage.Replicate.Endpoint,
		"${R2_ENDPOINT} must be expanded in storage.replicate.endpoint")
	assert.Equal(t, "test-access-key", cfg.Storage.Replicate.AccessKeyID,
		"${R2_ACCESS_KEY_ID} must be expanded in storage.replicate.access_key_id")
	assert.Equal(t, "test-secret-key", cfg.Storage.Replicate.SecretAccessKey,
		"${R2_SECRET_ACCESS_KEY} must be expanded in storage.replicate.secret_access_key")
}

// Regression test: when a config file IS provided, its storage.db wins over the
// env var. Previously env vars stomped per-agent config files, breaking
// multi-agent deployments.
func TestLoadWithEnvOverrides_ConfigFileWinsOverEnvVar(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("storage:\n  db: /from/config.db\n"), 0o644))

	t.Setenv("STATEPRIM_DB", "/from/env.db")

	cfg, err := LoadWithEnvOverrides(path, "STATEPRIM")
	require.NoError(t, err)
	assert.Equal(t, "/from/config.db", cfg.Storage.DB,
		"config file storage.db must win over STATEPRIM_DB env var when --config is provided")
}

// Regression test: when NO config file is provided, the env var is still used.
// This validates that the no-config / container deployment path remains intact.
func TestLoadWithEnvOverrides_EnvVarAppliedWithoutConfigFile(t *testing.T) {
	t.Setenv("STATEPRIM_DB", "/from/env.db")

	cfg, err := LoadWithEnvOverrides("", "STATEPRIM")
	require.NoError(t, err)
	assert.Equal(t, "/from/env.db", cfg.Storage.DB,
		"STATEPRIM_DB env var must be used when no config file is provided")
}

// Regression test: multi-agent scenario — two agents each have their own config
// file pointing to different databases. A global env var in the environment must
// NOT override either agent's per-instance config.
func TestLoadWithEnvOverrides_MultiAgentConfigFilesNotStompedByGlobalEnvVar(t *testing.T) {
	dir := t.TempDir()

	agentAConfig := filepath.Join(dir, "agent-a.yaml")
	agentBConfig := filepath.Join(dir, "agent-b.yaml")
	require.NoError(t, os.WriteFile(agentAConfig, []byte("storage:\n  db: /data/agent-a.db\n"), 0o644))
	require.NoError(t, os.WriteFile(agentBConfig, []byte("storage:\n  db: /data/agent-b.db\n"), 0o644))

	// Simulate a global env var present in the deployment environment.
	t.Setenv("TASKPRIM_DB", "/data/global-override.db")

	cfgA, err := LoadWithEnvOverrides(agentAConfig, "TASKPRIM")
	require.NoError(t, err)

	cfgB, err := LoadWithEnvOverrides(agentBConfig, "TASKPRIM")
	require.NoError(t, err)

	assert.Equal(t, "/data/agent-a.db", cfgA.Storage.DB,
		"agent A config file must win over global TASKPRIM_DB env var")
	assert.Equal(t, "/data/agent-b.db", cfgB.Storage.DB,
		"agent B config file must win over global TASKPRIM_DB env var")
	assert.NotEqual(t, cfgA.Storage.DB, cfgB.Storage.DB,
		"each agent must use its own database")
}
