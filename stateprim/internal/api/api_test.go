package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/propifly/primkit/primkit/db"
	"github.com/propifly/primkit/stateprim/internal/model"
	"github.com/propifly/primkit/stateprim/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestHandler creates an API handler backed by an in-memory SQLite store.
func newTestHandler(t *testing.T) (*Handler, store.Store) {
	t.Helper()
	database, err := db.OpenInMemory()
	require.NoError(t, err)
	s, err := store.NewFromDB(database)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := New(s, logger)
	return h, s
}

func doRequest(t *testing.T, h *Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(data)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	h.Router().ServeHTTP(rr, req)
	return rr
}

func decodeBody(t *testing.T, rr *httptest.ResponseRecorder, target interface{}) {
	t.Helper()
	require.NoError(t, json.NewDecoder(rr.Body).Decode(target))
}

// ---------------------------------------------------------------------------
// Set
// ---------------------------------------------------------------------------

func TestHandleSet(t *testing.T) {
	h, _ := newTestHandler(t)
	body := map[string]interface{}{
		"namespace": "config",
		"key":       "theme",
		"value":     "dark",
	}
	rr := doRequest(t, h, "POST", "/v1/records", body)
	assert.Equal(t, http.StatusOK, rr.Code)

	var r model.Record
	decodeBody(t, rr, &r)
	assert.Equal(t, "config", r.Namespace)
	assert.Equal(t, "theme", r.Key)
}

// ---------------------------------------------------------------------------
// Get
// ---------------------------------------------------------------------------

func TestHandleGet(t *testing.T) {
	h, s := newTestHandler(t)
	// Seed a record.
	doRequest(t, h, "POST", "/v1/records", map[string]interface{}{
		"namespace": "config", "key": "theme", "value": "dark",
	})

	rr := doRequest(t, h, "GET", "/v1/records/config/theme", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var r model.Record
	decodeBody(t, rr, &r)
	assert.Equal(t, "config", r.Namespace)
	_ = s // keep store alive
}

func TestHandleGet_NotFound(t *testing.T) {
	h, _ := newTestHandler(t)
	rr := doRequest(t, h, "GET", "/v1/records/ns/missing", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestHandleDelete(t *testing.T) {
	h, _ := newTestHandler(t)
	doRequest(t, h, "POST", "/v1/records", map[string]interface{}{
		"namespace": "ns", "key": "k", "value": "{}",
	})

	rr := doRequest(t, h, "DELETE", "/v1/records/ns/k", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify deleted.
	rr = doRequest(t, h, "GET", "/v1/records/ns/k", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandleDelete_NotFound(t *testing.T) {
	h, _ := newTestHandler(t)
	rr := doRequest(t, h, "DELETE", "/v1/records/ns/missing", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ---------------------------------------------------------------------------
// SetIfNew
// ---------------------------------------------------------------------------

func TestHandleSetIfNew(t *testing.T) {
	h, _ := newTestHandler(t)
	body := map[string]interface{}{"key": "msg:abc", "value": `{"sent":true}`}

	rr := doRequest(t, h, "POST", "/v1/records/dedup/set-if-new", body)
	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestHandleSetIfNew_AlreadyExists(t *testing.T) {
	h, _ := newTestHandler(t)
	body := map[string]interface{}{"key": "msg:abc", "value": `{"sent":true}`}

	doRequest(t, h, "POST", "/v1/records/dedup/set-if-new", body)
	rr := doRequest(t, h, "POST", "/v1/records/dedup/set-if-new", body)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

// ---------------------------------------------------------------------------
// Append
// ---------------------------------------------------------------------------

func TestHandleAppend(t *testing.T) {
	h, _ := newTestHandler(t)
	body := map[string]interface{}{"value": `{"action":"login"}`}

	rr := doRequest(t, h, "POST", "/v1/records/audit/append", body)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var r model.Record
	decodeBody(t, rr, &r)
	assert.Equal(t, "audit", r.Namespace)
	assert.True(t, r.Immutable)
	assert.Contains(t, r.Key, "a_")
}

// ---------------------------------------------------------------------------
// Has
// ---------------------------------------------------------------------------

func TestHandleHas_Exists(t *testing.T) {
	h, _ := newTestHandler(t)
	doRequest(t, h, "POST", "/v1/records", map[string]interface{}{
		"namespace": "ns", "key": "k", "value": "{}",
	})

	rr := doRequest(t, h, "GET", "/v1/records/ns/has/k", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var result map[string]bool
	decodeBody(t, rr, &result)
	assert.True(t, result["exists"])
}

func TestHandleHas_NotExists(t *testing.T) {
	h, _ := newTestHandler(t)
	rr := doRequest(t, h, "GET", "/v1/records/ns/has/missing", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var result map[string]bool
	decodeBody(t, rr, &result)
	assert.False(t, result["exists"])
}

// ---------------------------------------------------------------------------
// Query
// ---------------------------------------------------------------------------

func TestHandleQuery(t *testing.T) {
	h, _ := newTestHandler(t)
	doRequest(t, h, "POST", "/v1/records", map[string]interface{}{"namespace": "ns", "key": "k1", "value": "{}"})
	doRequest(t, h, "POST", "/v1/records", map[string]interface{}{"namespace": "ns", "key": "k2", "value": "{}"})

	rr := doRequest(t, h, "GET", "/v1/records/ns", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var records []*model.Record
	decodeBody(t, rr, &records)
	assert.Len(t, records, 2)
}

func TestHandleQuery_CountOnly(t *testing.T) {
	h, _ := newTestHandler(t)
	doRequest(t, h, "POST", "/v1/records", map[string]interface{}{"namespace": "ns", "key": "k1", "value": "{}"})

	rr := doRequest(t, h, "GET", "/v1/records/ns?count_only=true", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var result map[string]int
	decodeBody(t, rr, &result)
	assert.Equal(t, 1, result["count"])
}

func TestHandleQuery_WithPrefix(t *testing.T) {
	h, _ := newTestHandler(t)
	doRequest(t, h, "POST", "/v1/records", map[string]interface{}{"namespace": "ns", "key": "cfg:a", "value": "{}"})
	doRequest(t, h, "POST", "/v1/records", map[string]interface{}{"namespace": "ns", "key": "cfg:b", "value": "{}"})
	doRequest(t, h, "POST", "/v1/records", map[string]interface{}{"namespace": "ns", "key": "user:x", "value": "{}"})

	rr := doRequest(t, h, "GET", "/v1/records/ns?prefix=cfg:", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var records []*model.Record
	decodeBody(t, rr, &records)
	assert.Len(t, records, 2)
}

// ---------------------------------------------------------------------------
// Purge
// ---------------------------------------------------------------------------

func TestHandlePurge(t *testing.T) {
	h, _ := newTestHandler(t)
	rr := doRequest(t, h, "POST", "/v1/records/ns/purge", map[string]interface{}{"older_than": "24h"})
	assert.Equal(t, http.StatusOK, rr.Code)

	var result map[string]int
	decodeBody(t, rr, &result)
	assert.Equal(t, 0, result["deleted"])
}

// ---------------------------------------------------------------------------
// Namespaces
// ---------------------------------------------------------------------------

func TestHandleNamespaces(t *testing.T) {
	h, _ := newTestHandler(t)
	doRequest(t, h, "POST", "/v1/records", map[string]interface{}{"namespace": "ns1", "key": "k", "value": "{}"})
	doRequest(t, h, "POST", "/v1/records", map[string]interface{}{"namespace": "ns2", "key": "k", "value": "{}"})

	rr := doRequest(t, h, "GET", "/v1/namespaces", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var nss []model.NamespaceInfo
	decodeBody(t, rr, &nss)
	assert.Len(t, nss, 2)
}

// ---------------------------------------------------------------------------
// Stats
// ---------------------------------------------------------------------------

func TestHandleStats(t *testing.T) {
	h, _ := newTestHandler(t)
	doRequest(t, h, "POST", "/v1/records", map[string]interface{}{"namespace": "ns", "key": "k", "value": "{}"})

	rr := doRequest(t, h, "GET", "/v1/stats", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var stats model.Stats
	decodeBody(t, rr, &stats)
	assert.Equal(t, 1, stats.TotalRecords)
	assert.Equal(t, 1, stats.TotalNamespaces)
}

// ---------------------------------------------------------------------------
// Export / Import
// ---------------------------------------------------------------------------

func TestHandleExport(t *testing.T) {
	h, _ := newTestHandler(t)
	doRequest(t, h, "POST", "/v1/records", map[string]interface{}{"namespace": "ns", "key": "k1", "value": "{}"})
	doRequest(t, h, "POST", "/v1/records", map[string]interface{}{"namespace": "ns", "key": "k2", "value": "{}"})

	rr := doRequest(t, h, "GET", "/v1/export", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	var records []*model.Record
	decodeBody(t, rr, &records)
	assert.Len(t, records, 2)
}

func TestHandleImport(t *testing.T) {
	h, _ := newTestHandler(t)
	records := []map[string]interface{}{
		{"namespace": "imp", "key": "k1", "value": json.RawMessage(`{"a":1}`), "immutable": false, "created_at": "2026-03-03T10:00:00Z", "updated_at": "2026-03-03T10:00:00Z"},
	}

	rr := doRequest(t, h, "POST", "/v1/import", records)
	assert.Equal(t, http.StatusOK, rr.Code)

	var result map[string]int
	decodeBody(t, rr, &result)
	assert.Equal(t, 1, result["imported"])
}

// ---------------------------------------------------------------------------
// Error format
// ---------------------------------------------------------------------------

func TestErrorFormat(t *testing.T) {
	h, _ := newTestHandler(t)
	rr := doRequest(t, h, "GET", "/v1/records/ns/missing", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)

	var errResp struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	decodeBody(t, rr, &errResp)
	assert.Equal(t, "NOT_FOUND", errResp.Code)
}

// ---------------------------------------------------------------------------
// Integration
// ---------------------------------------------------------------------------

func TestLifecycle_SetGetDeleteViaAPI(t *testing.T) {
	h, _ := newTestHandler(t)

	// Set.
	doRequest(t, h, "POST", "/v1/records", map[string]interface{}{
		"namespace": "config", "key": "theme", "value": "light",
	})

	// Get.
	rr := doRequest(t, h, "GET", "/v1/records/config/theme", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Update.
	doRequest(t, h, "POST", "/v1/records", map[string]interface{}{
		"namespace": "config", "key": "theme", "value": "dark",
	})

	rr = doRequest(t, h, "GET", "/v1/records/config/theme", nil)
	var r model.Record
	decodeBody(t, rr, &r)
	assert.Equal(t, `"dark"`, string(r.Value))

	// Delete.
	rr = doRequest(t, h, "DELETE", "/v1/records/config/theme", nil)
	assert.Equal(t, http.StatusOK, rr.Code)

	rr = doRequest(t, h, "GET", "/v1/records/config/theme", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}
