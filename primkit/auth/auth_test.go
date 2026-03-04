package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/propifly/primkit/primkit/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_OpenMode(t *testing.T) {
	// No keys configured — all requests should pass through.
	v := NewValidator(nil)
	handler := v.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code, "open mode should allow all requests")
}

func TestValidator_ValidKey(t *testing.T) {
	keys := []KeyEntry{
		{Key: "tp_sk_test123", Name: "johanna"},
		{Key: "tp_sk_test456", Name: "andres"},
	}
	v := NewValidator(keys)

	var capturedName string
	handler := v.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedName = KeyNameFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer tp_sk_test123")
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "johanna", capturedName,
		"key name should be available in context")
}

func TestValidator_InvalidKey(t *testing.T) {
	v := NewValidator([]KeyEntry{{Key: "tp_sk_correct", Name: "test"}})
	handler := v.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called for invalid key")
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer tp_sk_wrong")
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp server.ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
	assert.Equal(t, "INVALID_KEY", errResp.Code)
}

func TestValidator_MissingHeader(t *testing.T) {
	v := NewValidator([]KeyEntry{{Key: "tp_sk_test", Name: "test"}})
	handler := v.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called without auth header")
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	// No Authorization header set.
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp server.ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
	assert.Equal(t, "AUTH_REQUIRED", errResp.Code)
}

func TestValidator_MalformedHeader(t *testing.T) {
	v := NewValidator([]KeyEntry{{Key: "tp_sk_test", Name: "test"}})
	handler := v.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called with malformed header")
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Basic dXNlcjpwYXNz") // Basic auth, not Bearer
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestValidator_SecondKey(t *testing.T) {
	keys := []KeyEntry{
		{Key: "tp_sk_first", Name: "agent1"},
		{Key: "tp_sk_second", Name: "agent2"},
	}
	v := NewValidator(keys)

	var capturedName string
	handler := v.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedName = KeyNameFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer tp_sk_second")
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "agent2", capturedName,
		"should match the second key in the list")
}

func TestKeyNameFromContext_NoAuth(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	name := KeyNameFromContext(r.Context())
	assert.Empty(t, name, "should return empty string when no auth in context")
}
