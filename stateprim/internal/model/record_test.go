package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecord_Validate_Valid(t *testing.T) {
	r := &Record{
		Namespace: "johanna/alerts",
		Key:       "thread_abc",
		Value:     json.RawMessage(`{"status":"ok"}`),
	}
	require.NoError(t, r.Validate())
}

func TestRecord_Validate_MissingNamespace(t *testing.T) {
	r := &Record{Key: "k", Value: json.RawMessage(`{}`)}
	err := r.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "namespace")
}

func TestRecord_Validate_MissingKey(t *testing.T) {
	r := &Record{Namespace: "ns", Value: json.RawMessage(`{}`)}
	err := r.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key")
}

func TestRecord_Validate_MissingValue(t *testing.T) {
	r := &Record{Namespace: "ns", Key: "k"}
	err := r.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "value")
}
