package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// --------------------------------------------------------------------
// Golden: add
// --------------------------------------------------------------------

func TestGolden_Add(t *testing.T) {
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
			out, err := execCmd(t, s, "add", "Test golden task", "--source", "test", "-f", tt.format)
			require.NoError(t, err)
			assertGolden(t, "add_"+tt.name, out)
		})
	}
}

// --------------------------------------------------------------------
// Golden: list
// --------------------------------------------------------------------

func TestGolden_List(t *testing.T) {
	tests := []struct {
		name   string
		format string
		seed   bool
	}{
		{name: "empty", format: "table", seed: false},
		{name: "table", format: "table", seed: true},
		{name: "json", format: "json", seed: true},
		{name: "quiet", format: "quiet", seed: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore(t)
			if tt.seed {
				seedTask(t, s, "Buy groceries", "inbox", "test", "urgent")
			}
			out, err := execCmd(t, s, "list", "-f", tt.format)
			require.NoError(t, err)
			assertGolden(t, "list_"+tt.name, out)
		})
	}
}

// --------------------------------------------------------------------
// Golden: get
// --------------------------------------------------------------------

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
			out, _ := execCmd(t, s, "add", "Get test task", "--source", "test", "-f", "quiet")
			id := strings.TrimSpace(out)
			out, err := execCmd(t, s, "get", id, "-f", tt.format)
			require.NoError(t, err)
			assertGolden(t, "get_"+tt.name, out)
		})
	}
}

// --------------------------------------------------------------------
// Golden: done
// --------------------------------------------------------------------

func TestGolden_Done(t *testing.T) {
	s := newTestStore(t)
	out, _ := execCmd(t, s, "add", "Done test task", "--source", "test", "-f", "quiet")
	id := strings.TrimSpace(out)
	out, err := execCmd(t, s, "done", id)
	require.NoError(t, err)
	assertGolden(t, "done_table", out)
}

// --------------------------------------------------------------------
// Golden: kill
// --------------------------------------------------------------------

func TestGolden_Kill(t *testing.T) {
	s := newTestStore(t)
	out, _ := execCmd(t, s, "add", "Kill test task", "--source", "test", "-f", "quiet")
	id := strings.TrimSpace(out)
	out, err := execCmd(t, s, "kill", id, "--reason", "not needed")
	require.NoError(t, err)
	assertGolden(t, "kill_table", out)
}

// --------------------------------------------------------------------
// Golden: labels
// --------------------------------------------------------------------

func TestGolden_Labels(t *testing.T) {
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
				seedTask(t, s, "Buy groceries", "inbox", "test", "urgent")
			}
			out, err := execCmd(t, s, "labels", "-f", tt.format)
			require.NoError(t, err)
			assertGolden(t, "labels_"+tt.name, out)
		})
	}
}

// --------------------------------------------------------------------
// Golden: lists
// --------------------------------------------------------------------

func TestGolden_Lists(t *testing.T) {
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
				seedTask(t, s, "Buy groceries", "inbox", "test", "urgent")
			}
			out, err := execCmd(t, s, "lists", "-f", tt.format)
			require.NoError(t, err)
			assertGolden(t, "lists_"+tt.name, out)
		})
	}
}

// --------------------------------------------------------------------
// Golden: stats
// --------------------------------------------------------------------

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
				seedTask(t, s, "Buy groceries", "inbox", "test", "urgent")
			}
			out, err := execCmd(t, s, "stats", "-f", tt.format)
			require.NoError(t, err)
			assertGolden(t, "stats_"+tt.name, out)
		})
	}
}
