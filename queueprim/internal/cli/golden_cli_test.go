package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// enqueue
// ---------------------------------------------------------------------------

func TestGolden_Enqueue(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{name: "table", format: "table"},
		{name: "json", format: "json"},
		{name: "quiet", format: "quiet"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore(t)
			out, err := execCmd(t, s, "enqueue", "emails", `{"to":"test@example.com"}`, "-f", tt.format)
			require.NoError(t, err)
			assertGolden(t, "enqueue_"+tt.name, out)
		})
	}
}

// ---------------------------------------------------------------------------
// list
// ---------------------------------------------------------------------------

func TestGolden_List(t *testing.T) {
	tests := []struct {
		name   string
		format string
		seed   bool
	}{
		{name: "empty", format: "table", seed: false},
		{name: "table", format: "table", seed: true},
		{name: "json", format: "json", seed: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore(t)
			if tt.seed {
				seedJob(t, s, "emails", map[string]string{"to": "test@example.com"})
			}
			out, err := execCmd(t, s, "list", "-f", tt.format)
			require.NoError(t, err)
			assertGolden(t, "list_"+tt.name, out)
		})
	}
}

// ---------------------------------------------------------------------------
// get
// ---------------------------------------------------------------------------

func TestGolden_Get(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{name: "table", format: "table"},
		{name: "json", format: "json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore(t)
			out, err := execCmd(t, s, "enqueue", "emails", `{"to":"test@example.com"}`, "-f", "quiet")
			require.NoError(t, err)
			id := strings.TrimSpace(out)
			out, err = execCmd(t, s, "get", id, "-f", tt.format)
			require.NoError(t, err)
			assertGolden(t, "get_"+tt.name, out)
		})
	}
}

// ---------------------------------------------------------------------------
// queues
// ---------------------------------------------------------------------------

func TestGolden_Queues(t *testing.T) {
	tests := []struct {
		name   string
		format string
		seed   bool
	}{
		{name: "empty", format: "table", seed: false},
		{name: "table", format: "table", seed: true},
		{name: "json", format: "json", seed: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore(t)
			if tt.seed {
				seedJob(t, s, "emails", map[string]string{"to": "test@example.com"})
			}
			out, err := execCmd(t, s, "queues", "-f", tt.format)
			require.NoError(t, err)
			assertGolden(t, "queues_"+tt.name, out)
		})
	}
}

// ---------------------------------------------------------------------------
// stats
// ---------------------------------------------------------------------------

func TestGolden_Stats(t *testing.T) {
	tests := []struct {
		name   string
		format string
		seed   bool
	}{
		{name: "empty", format: "table", seed: false},
		{name: "table", format: "table", seed: true},
		{name: "json", format: "json", seed: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore(t)
			if tt.seed {
				seedJob(t, s, "emails", map[string]string{"to": "test@example.com"})
			}
			out, err := execCmd(t, s, "stats", "-f", tt.format)
			require.NoError(t, err)
			assertGolden(t, "stats_"+tt.name, out)
		})
	}
}
