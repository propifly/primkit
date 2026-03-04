package store

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/propifly/primkit/primkit/db"
	"github.com/propifly/primkit/stateprim/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestStore creates a fresh in-memory store for each test. The store is
// automatically closed when the test ends.
func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	database, err := db.OpenInMemory()
	require.NoError(t, err)
	store, err := NewFromDB(database)
	require.NoError(t, err)
	t.Cleanup(func() { store.Close() })
	return store
}

// seedRecord inserts a record with sensible defaults. Override fields as needed.
func seedRecord(t *testing.T, s *SQLiteStore, overrides ...func(*model.Record)) *model.Record {
	t.Helper()
	r := &model.Record{
		Namespace: "test-ns",
		Key:       "test-key",
		Value:     json.RawMessage(`{"hello":"world"}`),
	}
	for _, fn := range overrides {
		fn(r)
	}
	require.NoError(t, s.Set(context.Background(), r))
	return r
}

// ---------------------------------------------------------------------------
// Set (upsert)
// ---------------------------------------------------------------------------

func TestSet_NewRecord(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	r := &model.Record{
		Namespace: "config",
		Key:       "theme",
		Value:     json.RawMessage(`"dark"`),
	}
	err := s.Set(ctx, r)
	require.NoError(t, err)

	assert.False(t, r.CreatedAt.IsZero(), "created_at should be populated")
	assert.False(t, r.UpdatedAt.IsZero(), "updated_at should be populated")

	got, err := s.Get(ctx, "config", "theme")
	require.NoError(t, err)
	assert.Equal(t, `"dark"`, string(got.Value))
	assert.False(t, got.Immutable)
}

func TestSet_UpdateExisting(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	seedRecord(t, s, func(r *model.Record) {
		r.Namespace = "config"
		r.Key = "theme"
		r.Value = json.RawMessage(`"light"`)
	})

	// Update the same key.
	r := &model.Record{
		Namespace: "config",
		Key:       "theme",
		Value:     json.RawMessage(`"dark"`),
	}
	err := s.Set(ctx, r)
	require.NoError(t, err)

	got, err := s.Get(ctx, "config", "theme")
	require.NoError(t, err)
	assert.Equal(t, `"dark"`, string(got.Value))
}

func TestSet_ImmutableRecord_RejectsUpdate(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Create an immutable record via Append.
	appended, err := s.Append(ctx, "log", []byte(`{"event":"started"}`))
	require.NoError(t, err)

	// Trying to Set over an immutable record should fail.
	r := &model.Record{
		Namespace: "log",
		Key:       appended.Key,
		Value:     json.RawMessage(`{"event":"stopped"}`),
	}
	err = s.Set(ctx, r)
	assert.ErrorIs(t, err, ErrImmutable)
}

func TestSet_Validation(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	err := s.Set(ctx, &model.Record{Key: "k", Value: json.RawMessage(`{}`)})
	assert.Error(t, err, "missing namespace should fail")

	err = s.Set(ctx, &model.Record{Namespace: "ns", Value: json.RawMessage(`{}`)})
	assert.Error(t, err, "missing key should fail")

	err = s.Set(ctx, &model.Record{Namespace: "ns", Key: "k"})
	assert.Error(t, err, "missing value should fail")
}

// ---------------------------------------------------------------------------
// Get
// ---------------------------------------------------------------------------

func TestGet_Existing(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	seedRecord(t, s, func(r *model.Record) {
		r.Namespace = "ns1"
		r.Key = "key1"
		r.Value = json.RawMessage(`{"count":42}`)
	})

	got, err := s.Get(ctx, "ns1", "key1")
	require.NoError(t, err)
	assert.Equal(t, "ns1", got.Namespace)
	assert.Equal(t, "key1", got.Key)
	assert.Equal(t, `{"count":42}`, string(got.Value))
}

func TestGet_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Get(context.Background(), "ns", "missing")
	assert.ErrorIs(t, err, ErrNotFound)
}

// ---------------------------------------------------------------------------
// Has
// ---------------------------------------------------------------------------

func TestHas_Existing(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	seedRecord(t, s)

	exists, err := s.Has(ctx, "test-ns", "test-key")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestHas_NotExisting(t *testing.T) {
	s := newTestStore(t)
	exists, err := s.Has(context.Background(), "ns", "nope")
	require.NoError(t, err)
	assert.False(t, exists)
}

// ---------------------------------------------------------------------------
// SetIfNew
// ---------------------------------------------------------------------------

func TestSetIfNew_CreatesNewRecord(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	r := &model.Record{
		Namespace: "dedup",
		Key:       "email:alice@example.com",
		Value:     json.RawMessage(`{"sent":true}`),
	}
	err := s.SetIfNew(ctx, r)
	require.NoError(t, err)
	assert.False(t, r.CreatedAt.IsZero())

	got, err := s.Get(ctx, "dedup", "email:alice@example.com")
	require.NoError(t, err)
	assert.Equal(t, `{"sent":true}`, string(got.Value))
}

func TestSetIfNew_RejectsExisting(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	r := &model.Record{
		Namespace: "dedup",
		Key:       "email:alice@example.com",
		Value:     json.RawMessage(`{"sent":true}`),
	}
	require.NoError(t, s.SetIfNew(ctx, r))

	// Second attempt should fail.
	r2 := &model.Record{
		Namespace: "dedup",
		Key:       "email:alice@example.com",
		Value:     json.RawMessage(`{"sent":false}`),
	}
	err := s.SetIfNew(ctx, r2)
	assert.ErrorIs(t, err, ErrAlreadyExists)

	// Original value should be unchanged.
	got, err := s.Get(ctx, "dedup", "email:alice@example.com")
	require.NoError(t, err)
	assert.Equal(t, `{"sent":true}`, string(got.Value))
}

func TestSetIfNew_Validation(t *testing.T) {
	s := newTestStore(t)
	err := s.SetIfNew(context.Background(), &model.Record{Key: "k", Value: json.RawMessage(`{}`)})
	assert.Error(t, err, "missing namespace should fail")
}

// ---------------------------------------------------------------------------
// Append
// ---------------------------------------------------------------------------

func TestAppend_CreatesImmutableRecord(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	r, err := s.Append(ctx, "events", []byte(`{"type":"login","user":"alice"}`))
	require.NoError(t, err)

	assert.Equal(t, "events", r.Namespace)
	assert.Contains(t, r.Key, "a_", "key should have append prefix")
	assert.True(t, r.Immutable)
	assert.False(t, r.CreatedAt.IsZero())

	// Verify it's persisted.
	got, err := s.Get(ctx, "events", r.Key)
	require.NoError(t, err)
	assert.True(t, got.Immutable)
}

func TestAppend_MultipleRecords_UniqueKeys(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	r1, err := s.Append(ctx, "events", []byte(`{"seq":1}`))
	require.NoError(t, err)
	r2, err := s.Append(ctx, "events", []byte(`{"seq":2}`))
	require.NoError(t, err)

	assert.NotEqual(t, r1.Key, r2.Key, "append keys must be unique")
}

func TestAppend_EmptyNamespace_Fails(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Append(context.Background(), "", []byte(`{}`))
	assert.Error(t, err)
}

func TestAppend_InvalidJSON_Fails(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Append(context.Background(), "events", []byte(`not json`))
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestDelete_Existing(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	seedRecord(t, s)

	err := s.Delete(ctx, "test-ns", "test-key")
	require.NoError(t, err)

	_, err = s.Get(ctx, "test-ns", "test-key")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDelete_NotFound(t *testing.T) {
	s := newTestStore(t)
	err := s.Delete(context.Background(), "ns", "missing")
	assert.ErrorIs(t, err, ErrNotFound)
}

// ---------------------------------------------------------------------------
// Query
// ---------------------------------------------------------------------------

func TestQuery_AllInNamespace(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		seedRecord(t, s, func(r *model.Record) {
			r.Key = "key-" + string(rune('a'+i))
		})
	}
	// Different namespace — should not appear.
	seedRecord(t, s, func(r *model.Record) {
		r.Namespace = "other"
		r.Key = "key-x"
	})

	records, count, err := s.Query(ctx, &model.QueryFilter{Namespace: "test-ns"})
	require.NoError(t, err)
	assert.Equal(t, 3, count)
	assert.Len(t, records, 3)
}

func TestQuery_KeyPrefix(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	seedRecord(t, s, func(r *model.Record) { r.Key = "cfg:theme" })
	seedRecord(t, s, func(r *model.Record) { r.Key = "cfg:lang" })
	seedRecord(t, s, func(r *model.Record) { r.Key = "user:name" })

	records, count, err := s.Query(ctx, &model.QueryFilter{
		Namespace: "test-ns",
		KeyPrefix: "cfg:",
	})
	require.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Len(t, records, 2)
}

func TestQuery_CountOnly(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		seedRecord(t, s, func(r *model.Record) {
			r.Key = "key-" + string(rune('a'+i))
		})
	}

	records, count, err := s.Query(ctx, &model.QueryFilter{
		Namespace: "test-ns",
		CountOnly: true,
	})
	require.NoError(t, err)
	assert.Nil(t, records)
	assert.Equal(t, 5, count)
}

func TestQuery_EmptyResult(t *testing.T) {
	s := newTestStore(t)
	records, count, err := s.Query(context.Background(), &model.QueryFilter{Namespace: "empty"})
	require.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, records)
	assert.NotNil(t, records, "should return empty slice, not nil")
}

func TestQuery_Since(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Create records — all will be very recent (within milliseconds).
	seedRecord(t, s, func(r *model.Record) { r.Key = "recent-1" })
	seedRecord(t, s, func(r *model.Record) { r.Key = "recent-2" })

	// Query for records updated within the last hour — should get all.
	records, count, err := s.Query(ctx, &model.QueryFilter{
		Namespace: "test-ns",
		Since:     time.Hour,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Len(t, records, 2)
}

// ---------------------------------------------------------------------------
// Purge
// ---------------------------------------------------------------------------

func TestPurge_RemovesOldRecords(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Manually insert a record with an old timestamp.
	oldTime := time.Now().UTC().Add(-48 * time.Hour).Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO records (namespace, key, value, immutable, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		"test-ns", "old-key", `{"old":true}`, false, oldTime, oldTime,
	)
	require.NoError(t, err)

	// Insert a recent record.
	seedRecord(t, s, func(r *model.Record) { r.Key = "fresh-key" })

	deleted, err := s.Purge(ctx, "test-ns", "24h")
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)

	// Old record gone, fresh one remains.
	_, err = s.Get(ctx, "test-ns", "old-key")
	assert.ErrorIs(t, err, ErrNotFound)

	_, err = s.Get(ctx, "test-ns", "fresh-key")
	assert.NoError(t, err)
}

func TestPurge_DayNotation(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Insert an old record.
	oldTime := time.Now().UTC().Add(-10 * 24 * time.Hour).Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO records (namespace, key, value, immutable, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		"ns", "old", `{}`, false, oldTime, oldTime,
	)
	require.NoError(t, err)

	deleted, err := s.Purge(ctx, "ns", "7d")
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)
}

func TestPurge_InvalidDuration(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Purge(context.Background(), "ns", "invalid")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ListNamespaces
// ---------------------------------------------------------------------------

func TestListNamespaces(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	seedRecord(t, s, func(r *model.Record) { r.Namespace = "alpha"; r.Key = "k1" })
	seedRecord(t, s, func(r *model.Record) { r.Namespace = "alpha"; r.Key = "k2" })
	seedRecord(t, s, func(r *model.Record) { r.Namespace = "beta"; r.Key = "k1" })

	nss, err := s.ListNamespaces(ctx)
	require.NoError(t, err)
	assert.Len(t, nss, 2)

	assert.Equal(t, "alpha", nss[0].Namespace)
	assert.Equal(t, 2, nss[0].Count)
	assert.Equal(t, "beta", nss[1].Namespace)
	assert.Equal(t, 1, nss[1].Count)
}

func TestListNamespaces_Empty(t *testing.T) {
	s := newTestStore(t)
	nss, err := s.ListNamespaces(context.Background())
	require.NoError(t, err)
	assert.Empty(t, nss)
	assert.NotNil(t, nss, "should return empty slice, not nil")
}

// ---------------------------------------------------------------------------
// Stats
// ---------------------------------------------------------------------------

func TestStats(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	seedRecord(t, s, func(r *model.Record) { r.Namespace = "ns1"; r.Key = "k1" })
	seedRecord(t, s, func(r *model.Record) { r.Namespace = "ns1"; r.Key = "k2" })
	seedRecord(t, s, func(r *model.Record) { r.Namespace = "ns2"; r.Key = "k1" })

	stats, err := s.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, stats.TotalRecords)
	assert.Equal(t, 2, stats.TotalNamespaces)
}

func TestStats_Empty(t *testing.T) {
	s := newTestStore(t)
	stats, err := s.Stats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, stats.TotalRecords)
	assert.Equal(t, 0, stats.TotalNamespaces)
}

// ---------------------------------------------------------------------------
// Export / Import
// ---------------------------------------------------------------------------

func TestExportRecords_All(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	seedRecord(t, s, func(r *model.Record) { r.Namespace = "ns1"; r.Key = "k1" })
	seedRecord(t, s, func(r *model.Record) { r.Namespace = "ns2"; r.Key = "k2" })

	records, err := s.ExportRecords(ctx, "")
	require.NoError(t, err)
	assert.Len(t, records, 2)
}

func TestExportRecords_FilteredByNamespace(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	seedRecord(t, s, func(r *model.Record) { r.Namespace = "ns1"; r.Key = "k1" })
	seedRecord(t, s, func(r *model.Record) { r.Namespace = "ns2"; r.Key = "k2" })

	records, err := s.ExportRecords(ctx, "ns1")
	require.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "ns1", records[0].Namespace)
}

func TestExportRecords_Empty(t *testing.T) {
	s := newTestStore(t)
	records, err := s.ExportRecords(context.Background(), "")
	require.NoError(t, err)
	assert.Empty(t, records)
	assert.NotNil(t, records)
}

func TestImportRecords(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	now := time.Now().UTC()
	records := []*model.Record{
		{Namespace: "imported", Key: "k1", Value: json.RawMessage(`{"a":1}`), CreatedAt: now, UpdatedAt: now},
		{Namespace: "imported", Key: "k2", Value: json.RawMessage(`{"b":2}`), CreatedAt: now, UpdatedAt: now},
	}
	err := s.ImportRecords(ctx, records)
	require.NoError(t, err)

	// Verify records are persisted.
	got, err := s.Get(ctx, "imported", "k1")
	require.NoError(t, err)
	assert.Equal(t, `{"a":1}`, string(got.Value))

	got, err = s.Get(ctx, "imported", "k2")
	require.NoError(t, err)
	assert.Equal(t, `{"b":2}`, string(got.Value))
}

func TestImportRecords_OverwritesExisting(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	seedRecord(t, s, func(r *model.Record) {
		r.Namespace = "ns"
		r.Key = "existing"
		r.Value = json.RawMessage(`{"v":1}`)
	})

	now := time.Now().UTC()
	err := s.ImportRecords(ctx, []*model.Record{
		{Namespace: "ns", Key: "existing", Value: json.RawMessage(`{"v":2}`), CreatedAt: now, UpdatedAt: now},
	})
	require.NoError(t, err)

	got, err := s.Get(ctx, "ns", "existing")
	require.NoError(t, err)
	assert.Equal(t, `{"v":2}`, string(got.Value))
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func TestParseDuration_GoStyle(t *testing.T) {
	d, err := parseDuration("2h30m")
	require.NoError(t, err)
	assert.Equal(t, 2*time.Hour+30*time.Minute, d)
}

func TestParseDuration_DayNotation(t *testing.T) {
	d, err := parseDuration("7d")
	require.NoError(t, err)
	assert.Equal(t, 7*24*time.Hour, d)
}

func TestParseDuration_Invalid(t *testing.T) {
	_, err := parseDuration("bogus")
	assert.Error(t, err)
}

func TestGenerateAppendKey_Format(t *testing.T) {
	now := time.Date(2026, 3, 3, 14, 30, 0, 0, time.UTC)
	key := generateAppendKey(now)

	assert.Contains(t, key, "a_20260303T143000_")
	assert.Len(t, key, len("a_20260303T143000_")+8) // 8 hex chars = 4 bytes
}

// ---------------------------------------------------------------------------
// Integration: full lifecycle
// ---------------------------------------------------------------------------

func TestLifecycle_KeyValueState(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// 1. Set a value.
	r := &model.Record{
		Namespace: "config",
		Key:       "app.theme",
		Value:     json.RawMessage(`"light"`),
	}
	require.NoError(t, s.Set(ctx, r))

	// 2. Read it back.
	got, err := s.Get(ctx, "config", "app.theme")
	require.NoError(t, err)
	assert.Equal(t, `"light"`, string(got.Value))

	// 3. Update it.
	r.Value = json.RawMessage(`"dark"`)
	require.NoError(t, s.Set(ctx, r))

	got, err = s.Get(ctx, "config", "app.theme")
	require.NoError(t, err)
	assert.Equal(t, `"dark"`, string(got.Value))

	// 4. Check existence.
	exists, err := s.Has(ctx, "config", "app.theme")
	require.NoError(t, err)
	assert.True(t, exists)

	// 5. Delete it.
	require.NoError(t, s.Delete(ctx, "config", "app.theme"))

	exists, err = s.Has(ctx, "config", "app.theme")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestLifecycle_DedupLookup(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// First attempt: success.
	r := &model.Record{
		Namespace: "sent-emails",
		Key:       "msg:abc123",
		Value:     json.RawMessage(`{"to":"alice@example.com"}`),
	}
	require.NoError(t, s.SetIfNew(ctx, r))

	// Second attempt: already exists.
	r2 := &model.Record{
		Namespace: "sent-emails",
		Key:       "msg:abc123",
		Value:     json.RawMessage(`{"to":"bob@example.com"}`),
	}
	err := s.SetIfNew(ctx, r2)
	assert.ErrorIs(t, err, ErrAlreadyExists)

	// Original value preserved.
	got, err := s.Get(ctx, "sent-emails", "msg:abc123")
	require.NoError(t, err)
	assert.Contains(t, string(got.Value), "alice")
}

func TestLifecycle_AppendLog(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Append several log entries.
	_, err := s.Append(ctx, "audit", []byte(`{"action":"login","user":"alice"}`))
	require.NoError(t, err)
	_, err = s.Append(ctx, "audit", []byte(`{"action":"update","user":"alice"}`))
	require.NoError(t, err)
	_, err = s.Append(ctx, "audit", []byte(`{"action":"logout","user":"alice"}`))
	require.NoError(t, err)

	// Query all.
	records, count, err := s.Query(ctx, &model.QueryFilter{Namespace: "audit"})
	require.NoError(t, err)
	assert.Equal(t, 3, count)
	assert.Len(t, records, 3)

	// All should be immutable.
	for _, rec := range records {
		assert.True(t, rec.Immutable)
	}

	// Count only.
	_, countOnly, err := s.Query(ctx, &model.QueryFilter{Namespace: "audit", CountOnly: true})
	require.NoError(t, err)
	assert.Equal(t, 3, countOnly)

	// Stats should reflect.
	stats, err := s.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, stats.TotalRecords)
	assert.Equal(t, 1, stats.TotalNamespaces)
}

func TestLifecycle_ExportImportRoundTrip(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Populate with data.
	seedRecord(t, s, func(r *model.Record) { r.Namespace = "ns1"; r.Key = "k1"; r.Value = json.RawMessage(`{"a":1}`) })
	seedRecord(t, s, func(r *model.Record) { r.Namespace = "ns2"; r.Key = "k2"; r.Value = json.RawMessage(`{"b":2}`) })

	// Export.
	exported, err := s.ExportRecords(ctx, "")
	require.NoError(t, err)
	assert.Len(t, exported, 2)

	// Import into a fresh store.
	s2 := newTestStore(t)
	require.NoError(t, s2.ImportRecords(ctx, exported))

	// Verify.
	got, err := s2.Get(ctx, "ns1", "k1")
	require.NoError(t, err)
	assert.Equal(t, `{"a":1}`, string(got.Value))

	got, err = s2.Get(ctx, "ns2", "k2")
	require.NoError(t, err)
	assert.Equal(t, `{"b":2}`, string(got.Value))
}
