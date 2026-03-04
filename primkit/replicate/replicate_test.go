package replicate

import (
	"log/slog"
	"os"
	"testing"

	"github.com/propifly/primkit/primkit/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	cfg := config.ReplicateConfig{
		Enabled:  true,
		Provider: "r2",
		Bucket:   "test-bucket",
		Path:     "test.db",
	}

	rep := New(cfg, nil)
	require.NotNil(t, rep)
	assert.Equal(t, "test-bucket", rep.cfg.Bucket)
	assert.Equal(t, "r2", rep.cfg.Provider)
	assert.False(t, rep.IsRunning())
}

func TestNew_WithLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	rep := New(config.ReplicateConfig{}, logger)
	require.NotNil(t, rep)
	assert.Equal(t, logger, rep.logger)
}

func TestNew_NilLogger_UsesDefault(t *testing.T) {
	rep := New(config.ReplicateConfig{}, nil)
	require.NotNil(t, rep)
	assert.NotNil(t, rep.logger)
}

func TestNewS3Client_R2(t *testing.T) {
	rep := New(config.ReplicateConfig{
		Provider:        "r2",
		Bucket:          "my-bucket",
		Path:            "db/taskprim.db",
		Endpoint:        "https://abc123.r2.cloudflarestorage.com",
		AccessKeyID:     "access-key",
		SecretAccessKey: "secret-key",
	}, nil)

	client := rep.newS3Client()
	assert.Equal(t, "my-bucket", client.Bucket)
	assert.Equal(t, "db/taskprim.db", client.Path)
	assert.Equal(t, "https://abc123.r2.cloudflarestorage.com", client.Endpoint)
	assert.Equal(t, "access-key", client.AccessKeyID)
	assert.Equal(t, "secret-key", client.SecretAccessKey)
	assert.Equal(t, "auto", client.Region)
}

func TestNewS3Client_S3(t *testing.T) {
	rep := New(config.ReplicateConfig{
		Provider:        "s3",
		Bucket:          "my-s3-bucket",
		Path:            "backups/db",
		AccessKeyID:     "aws-key",
		SecretAccessKey: "aws-secret",
	}, nil)

	client := rep.newS3Client()
	assert.Equal(t, "my-s3-bucket", client.Bucket)
	assert.Equal(t, "us-east-1", client.Region) // default for non-R2
	assert.Equal(t, "", client.Endpoint)
}

func TestNewS3Client_B2(t *testing.T) {
	rep := New(config.ReplicateConfig{
		Provider:        "b2",
		Bucket:          "b2-bucket",
		Endpoint:        "https://s3.us-west-001.backblazeb2.com",
		AccessKeyID:     "b2-key",
		SecretAccessKey: "b2-secret",
	}, nil)

	client := rep.newS3Client()
	assert.Equal(t, "b2-bucket", client.Bucket)
	assert.Equal(t, "us-east-1", client.Region) // default for non-R2
	assert.Equal(t, "https://s3.us-west-001.backblazeb2.com", client.Endpoint)
}

func TestRestoreIfNeeded_LocalDBExists(t *testing.T) {
	// Create a temporary file to simulate an existing DB.
	f, err := os.CreateTemp(t.TempDir(), "test-*.db")
	require.NoError(t, err)
	f.Close()

	rep := New(config.ReplicateConfig{
		Provider: "r2",
		Bucket:   "test-bucket",
	}, nil)

	// Should be a no-op when the file already exists.
	err = rep.RestoreIfNeeded(t.Context(), f.Name())
	assert.NoError(t, err)
}

func TestStop_NotRunning(t *testing.T) {
	rep := New(config.ReplicateConfig{}, nil)
	// Stop on a replicator that was never started should be a no-op.
	err := rep.Stop()
	assert.NoError(t, err)
}

func TestIsRunning(t *testing.T) {
	rep := New(config.ReplicateConfig{}, nil)
	assert.False(t, rep.IsRunning())
}
