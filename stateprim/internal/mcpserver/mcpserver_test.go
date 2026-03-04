package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/propifly/primkit/primkit/db"
	"github.com/propifly/primkit/stateprim/internal/model"
	"github.com/propifly/primkit/stateprim/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mcpsvr "github.com/mark3labs/mcp-go/server"
)

func newTestMCP(t *testing.T) (*mcpsvr.MCPServer, store.Store) {
	t.Helper()
	database, err := db.OpenInMemory()
	require.NoError(t, err)
	s, err := store.NewFromDB(database)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })

	srv := New(s, "test")
	return srv, s
}

// callTool sends a JSON-RPC tools/call request and returns the parsed response.
func callTool(t *testing.T, srv *mcpsvr.MCPServer, name string, args map[string]interface{}) json.RawMessage {
	t.Helper()
	argsJSON, err := json.Marshal(args)
	require.NoError(t, err)

	reqJSON := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": %q,
			"arguments": %s
		}
	}`, name, string(argsJSON))

	resp := srv.HandleMessage(context.Background(), []byte(reqJSON))
	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var envelope struct {
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
	}
	require.NoError(t, json.Unmarshal(data, &envelope))

	if envelope.Result.IsError {
		t.Logf("Tool error: %s", envelope.Result.Content[0].Text)
	}

	if len(envelope.Result.Content) == 0 {
		return nil
	}
	return json.RawMessage(envelope.Result.Content[0].Text)
}

// callToolJSON calls a tool and unmarshals the result text into target.
func callToolJSON(t *testing.T, srv *mcpsvr.MCPServer, name string, args map[string]interface{}, target interface{}) {
	t.Helper()
	raw := callTool(t, srv, name, args)
	require.NotNil(t, raw)
	require.NoError(t, json.Unmarshal(raw, target))
}

// ---------------------------------------------------------------------------
// Tools
// ---------------------------------------------------------------------------

func TestMCP_Set(t *testing.T) {
	srv, _ := newTestMCP(t)
	var r model.Record
	callToolJSON(t, srv, "stateprim_set", map[string]interface{}{
		"namespace": "config", "key": "theme", "value": `"dark"`,
	}, &r)
	assert.Equal(t, "config", r.Namespace)
	assert.Equal(t, "theme", r.Key)
}

func TestMCP_Get(t *testing.T) {
	srv, _ := newTestMCP(t)
	callTool(t, srv, "stateprim_set", map[string]interface{}{
		"namespace": "config", "key": "theme", "value": `"dark"`,
	})

	var r model.Record
	callToolJSON(t, srv, "stateprim_get", map[string]interface{}{
		"namespace": "config", "key": "theme",
	}, &r)
	assert.Equal(t, `"dark"`, string(r.Value))
}

func TestMCP_Has(t *testing.T) {
	srv, _ := newTestMCP(t)
	callTool(t, srv, "stateprim_set", map[string]interface{}{
		"namespace": "ns", "key": "k", "value": `{}`,
	})

	var result map[string]bool
	callToolJSON(t, srv, "stateprim_has", map[string]interface{}{
		"namespace": "ns", "key": "k",
	}, &result)
	assert.True(t, result["exists"])
}

func TestMCP_SetIfNew(t *testing.T) {
	srv, _ := newTestMCP(t)
	var r model.Record
	callToolJSON(t, srv, "stateprim_set_if_new", map[string]interface{}{
		"namespace": "dedup", "key": "msg:abc", "value": `{"sent":true}`,
	}, &r)
	assert.Equal(t, "dedup", r.Namespace)
}

func TestMCP_Append(t *testing.T) {
	srv, _ := newTestMCP(t)
	var r model.Record
	callToolJSON(t, srv, "stateprim_append", map[string]interface{}{
		"namespace": "audit", "value": `{"action":"login"}`,
	}, &r)
	assert.Equal(t, "audit", r.Namespace)
	assert.True(t, r.Immutable)
	assert.Contains(t, r.Key, "a_")
}

func TestMCP_Delete(t *testing.T) {
	srv, _ := newTestMCP(t)
	callTool(t, srv, "stateprim_set", map[string]interface{}{
		"namespace": "ns", "key": "k", "value": `{}`,
	})

	var result map[string]string
	callToolJSON(t, srv, "stateprim_delete", map[string]interface{}{
		"namespace": "ns", "key": "k",
	}, &result)
	assert.Equal(t, "ok", result["status"])
}

func TestMCP_Query(t *testing.T) {
	srv, _ := newTestMCP(t)
	callTool(t, srv, "stateprim_set", map[string]interface{}{"namespace": "ns", "key": "k1", "value": `{}`})
	callTool(t, srv, "stateprim_set", map[string]interface{}{"namespace": "ns", "key": "k2", "value": `{}`})

	var records []*model.Record
	callToolJSON(t, srv, "stateprim_query", map[string]interface{}{
		"namespace": "ns",
	}, &records)
	assert.Len(t, records, 2)
}

func TestMCP_Purge(t *testing.T) {
	srv, _ := newTestMCP(t)
	var result map[string]int
	callToolJSON(t, srv, "stateprim_purge", map[string]interface{}{
		"namespace": "ns", "older_than": "24h",
	}, &result)
	assert.Equal(t, 0, result["deleted"])
}

func TestMCP_Namespaces(t *testing.T) {
	srv, _ := newTestMCP(t)
	callTool(t, srv, "stateprim_set", map[string]interface{}{"namespace": "ns1", "key": "k", "value": `{}`})

	var nss []model.NamespaceInfo
	callToolJSON(t, srv, "stateprim_namespaces", map[string]interface{}{}, &nss)
	assert.Len(t, nss, 1)
	assert.Equal(t, "ns1", nss[0].Namespace)
}

func TestMCP_Stats(t *testing.T) {
	srv, _ := newTestMCP(t)
	callTool(t, srv, "stateprim_set", map[string]interface{}{"namespace": "ns", "key": "k", "value": `{}`})

	var stats model.Stats
	callToolJSON(t, srv, "stateprim_stats", map[string]interface{}{}, &stats)
	assert.Equal(t, 1, stats.TotalRecords)
	assert.Equal(t, 1, stats.TotalNamespaces)
}

// ---------------------------------------------------------------------------
// Tool listing
// ---------------------------------------------------------------------------

func TestMCP_ListTools(t *testing.T) {
	srv, _ := newTestMCP(t)

	reqJSON := `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`
	resp := srv.HandleMessage(context.Background(), []byte(reqJSON))
	data, _ := json.Marshal(resp)

	var envelope struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	require.NoError(t, json.Unmarshal(data, &envelope))
	assert.Len(t, envelope.Result.Tools, 10)

	names := make(map[string]bool)
	for _, tool := range envelope.Result.Tools {
		names[tool.Name] = true
	}
	for _, expected := range []string{
		"stateprim_set", "stateprim_get", "stateprim_has",
		"stateprim_set_if_new", "stateprim_append", "stateprim_delete",
		"stateprim_query", "stateprim_purge", "stateprim_namespaces",
		"stateprim_stats",
	} {
		assert.True(t, names[expected], "missing tool: %s", expected)
	}
}

// ---------------------------------------------------------------------------
// Integration
// ---------------------------------------------------------------------------

func TestMCP_Lifecycle(t *testing.T) {
	srv, _ := newTestMCP(t)

	// Set.
	callTool(t, srv, "stateprim_set", map[string]interface{}{
		"namespace": "config", "key": "theme", "value": `"light"`,
	})

	// Get.
	var r model.Record
	callToolJSON(t, srv, "stateprim_get", map[string]interface{}{
		"namespace": "config", "key": "theme",
	}, &r)
	assert.Equal(t, `"light"`, string(r.Value))

	// Update.
	callTool(t, srv, "stateprim_set", map[string]interface{}{
		"namespace": "config", "key": "theme", "value": `"dark"`,
	})
	callToolJSON(t, srv, "stateprim_get", map[string]interface{}{
		"namespace": "config", "key": "theme",
	}, &r)
	assert.Equal(t, `"dark"`, string(r.Value))

	// Delete.
	callTool(t, srv, "stateprim_delete", map[string]interface{}{
		"namespace": "config", "key": "theme",
	})

	// Has after delete.
	var hasResult map[string]bool
	callToolJSON(t, srv, "stateprim_has", map[string]interface{}{
		"namespace": "config", "key": "theme",
	}, &hasResult)
	assert.False(t, hasResult["exists"])
}
