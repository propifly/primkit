// Package replicate provides embedded Litestream replication for primkit
// databases. It wraps the Litestream library to continuously stream WAL
// frames to S3-compatible object storage (AWS S3, Cloudflare R2, Backblaze B2).
//
// Both taskprim and stateprim use this package. Replication runs in serve
// and MCP modes; CLI mode uses RestoreIfNeeded for auto-recovery.
package replicate

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/propifly/primkit/primkit/config"
	"github.com/benbjohnson/litestream"
	lss3 "github.com/benbjohnson/litestream/s3"
)

// Replicator manages Litestream replication for a single SQLite database.
// It handles restore-on-startup, continuous WAL streaming, and graceful
// shutdown with final sync.
type Replicator struct {
	cfg    config.ReplicateConfig
	logger *slog.Logger

	db    *litestream.DB
	store *litestream.Store
}

// New creates a Replicator from the given replication config. The replicator
// is inert until Start is called. If logger is nil, a default logger is used.
func New(cfg config.ReplicateConfig, logger *slog.Logger) *Replicator {
	if logger == nil {
		logger = slog.Default()
	}
	return &Replicator{
		cfg:    cfg,
		logger: logger,
	}
}

// newS3Client creates an S3 ReplicaClient from the replication config.
// Works with AWS S3, Cloudflare R2, Backblaze B2, and any S3-compatible service.
func (r *Replicator) newS3Client() *lss3.ReplicaClient {
	client := lss3.NewReplicaClient()
	client.Bucket = r.cfg.Bucket
	client.Path = r.cfg.Path
	client.AccessKeyID = r.cfg.AccessKeyID
	client.SecretAccessKey = r.cfg.SecretAccessKey

	if r.cfg.Endpoint != "" {
		client.Endpoint = r.cfg.Endpoint
		// ForcePathStyle is auto-enabled by Litestream when Endpoint is set.
	}

	// Set region based on provider. R2 uses "auto", others default to us-east-1.
	switch r.cfg.Provider {
	case "r2":
		client.Region = "auto"
	default:
		client.Region = lss3.DefaultRegion
	}

	return client
}

// RestoreIfNeeded restores the database from the replica if the local file
// does not exist. This is called before opening the database so that serve
// and MCP modes can recover from a previous deployment's backup.
//
// If the local file already exists, this is a no-op. If no backup exists
// on the remote, this is also a no-op (a fresh database will be created).
func (r *Replicator) RestoreIfNeeded(ctx context.Context, dbPath string) error {
	// Skip if local database already exists.
	if _, err := os.Stat(dbPath); err == nil {
		r.logger.Info("local database exists, skipping restore", "path", dbPath)
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking database file: %w", err)
	}

	r.logger.Info("local database not found, attempting restore from replica",
		"path", dbPath,
		"bucket", r.cfg.Bucket,
		"remote_path", r.cfg.Path,
	)

	return r.Restore(ctx, dbPath)
}

// Restore downloads the database from the replica to the given local path.
// Unlike RestoreIfNeeded, this always attempts the restore regardless of
// whether the local file exists (it will overwrite).
func (r *Replicator) Restore(ctx context.Context, dbPath string) error {
	db := litestream.NewDB(dbPath)
	db.Logger = r.logger.With("component", "litestream")

	client := r.newS3Client()
	replica := litestream.NewReplicaWithClient(db, client)
	db.Replica = replica

	if err := db.EnsureExists(ctx); err != nil {
		return fmt.Errorf("restoring from replica: %w", err)
	}

	r.logger.Info("restore complete", "path", dbPath)
	return db.Close(ctx)
}

// Start begins continuous WAL replication for the database at dbPath. The
// Litestream DB monitors the WAL file and streams changes to the remote
// replica. Call Stop to gracefully shut down replication.
//
// The caller should open their sql.DB AFTER calling Start, or at least
// ensure the database file exists before calling Start.
func (r *Replicator) Start(dbPath string) error {
	r.logger.Info("starting litestream replication",
		"path", dbPath,
		"provider", r.cfg.Provider,
		"bucket", r.cfg.Bucket,
		"remote_path", r.cfg.Path,
	)

	db := litestream.NewDB(dbPath)
	db.Logger = r.logger.With("component", "litestream")

	client := r.newS3Client()
	replica := litestream.NewReplicaWithClient(db, client)
	db.Replica = replica

	// Create the store which manages the replication lifecycle.
	// Level 0: raw LTX files from WAL monitoring.
	// Level 1: compacted every 10 seconds for faster restores.
	levels := litestream.CompactionLevels{
		{Level: 0},
		{Level: 1, Interval: 10 * time.Second},
	}
	store := litestream.NewStore([]*litestream.DB{db}, levels)

	if err := store.Open(context.Background()); err != nil {
		return fmt.Errorf("starting litestream store: %w", err)
	}

	r.db = db
	r.store = store

	r.logger.Info("litestream replication started")
	return nil
}

// Stop gracefully shuts down replication. It performs a final sync to ensure
// all pending WAL frames are uploaded, then closes the Litestream store.
func (r *Replicator) Stop() error {
	if r.store == nil {
		return nil
	}

	r.logger.Info("stopping litestream replication")

	// Final sync to push any remaining WAL changes.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := r.db.SyncAndWait(ctx); err != nil {
		r.logger.Warn("final sync failed", "error", err)
	}

	if err := r.store.Close(ctx); err != nil {
		return fmt.Errorf("closing litestream store: %w", err)
	}

	r.logger.Info("litestream replication stopped")
	r.db = nil
	r.store = nil
	return nil
}

// IsRunning returns true if replication is currently active.
func (r *Replicator) IsRunning() bool {
	return r.store != nil
}
