package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecode_ValidJSON(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	body := strings.NewReader(`{"name":"test","age":30}`)
	r := httptest.NewRequest(http.MethodPost, "/", body)

	var p payload
	err := Decode(r, &p)
	require.NoError(t, err)
	assert.Equal(t, "test", p.Name)
	assert.Equal(t, 30, p.Age)
}

func TestDecode_InvalidJSON(t *testing.T) {
	body := strings.NewReader(`{not valid}`)
	r := httptest.NewRequest(http.MethodPost, "/", body)

	var p struct{}
	err := Decode(r, &p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON")
}

func TestDecode_NilBody(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Body = nil

	var p struct{}
	err := Decode(r, &p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestDecode_UnknownFields(t *testing.T) {
	body := strings.NewReader(`{"name":"test","unknown_field":"value"}`)
	r := httptest.NewRequest(http.MethodPost, "/", body)

	var p struct {
		Name string `json:"name"`
	}
	err := Decode(r, &p)
	assert.Error(t, err, "unknown fields should be rejected")
}

func TestRespond(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"status": "ok"}

	Respond(w, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var result map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "ok", result["status"])
}

func TestRespondError(t *testing.T) {
	w := httptest.NewRecorder()
	RespondError(w, http.StatusNotFound, "NOT_FOUND", "task not found")

	assert.Equal(t, http.StatusNotFound, w.Code)

	var result ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "NOT_FOUND", result.Code)
	assert.Equal(t, "task not found", result.Error)
}

func TestChain(t *testing.T) {
	// Track the order middleware executes.
	var order []string

	mw := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+"-before")
				next.ServeHTTP(w, r)
				order = append(order, name+"-after")
			})
		}
	}

	handler := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "handler")
		}),
		mw("first"),
		mw("second"),
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(w, r)

	assert.Equal(t, []string{
		"first-before", "second-before", "handler", "second-after", "first-after",
	}, order, "middleware should execute in listed order")
}

func TestRequestID_Generated(t *testing.T) {
	handler := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := RequestIDFromContext(r.Context())
		assert.NotEmpty(t, id, "request ID should be set in context")
		io.WriteString(w, id)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(w, r)

	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestRequestID_Reused(t *testing.T) {
	handler := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := RequestIDFromContext(r.Context())
		io.WriteString(w, id)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Request-ID", "my-trace-id")
	handler.ServeHTTP(w, r)

	assert.Equal(t, "my-trace-id", w.Header().Get("X-Request-ID"),
		"existing X-Request-ID should be reused")
	assert.Equal(t, "my-trace-id", w.Body.String())
}

func TestLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	handler := Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/v1/tasks", nil)
	handler.ServeHTTP(w, r)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "POST")
	assert.Contains(t, logOutput, "/v1/tasks")
}

func TestRecovery(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	handler := Recovery(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something broke")
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code,
		"panic should be caught and return 500")

	var result ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "INTERNAL_ERROR", result.Code)
}
