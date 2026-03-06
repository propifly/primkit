package docgen

import (
	"strings"
	"testing"
)

func TestRenderCommandTable_empty(t *testing.T) {
	meta := PrimMeta{Name: "testprim", Commands: nil}
	got := RenderCommandTable(meta)
	if !strings.Contains(got, "| Command |") {
		t.Errorf("expected header row, got: %q", got)
	}
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines (header + separator), got %d", len(lines))
	}
}

func TestRenderCommandTable_singleCommand(t *testing.T) {
	meta := PrimMeta{
		Name: "testprim",
		Commands: []CmdMeta{
			{
				Name:     "enqueue",
				Synopsis: "enqueue <queue> <json_payload>",
				Short:    "Enqueue a new job",
				Flags: []FlagMeta{
					{Name: "priority", Usage: "priority: high, normal, or low", Default: "normal"},
					{Name: "type", Usage: "job type category", Default: ""},
				},
			},
		},
	}
	got := RenderCommandTable(meta)

	if !strings.Contains(got, "`enqueue`") {
		t.Errorf("expected command name in output, got: %q", got)
	}
	if !strings.Contains(got, "enqueue <queue> <json_payload>") {
		t.Errorf("expected synopsis in output, got: %q", got)
	}
	if !strings.Contains(got, "`--priority`") {
		t.Errorf("expected priority flag in output, got: %q", got)
	}
	if !strings.Contains(got, "default: `normal`") {
		t.Errorf("expected default value annotation, got: %q", got)
	}
	// --type has empty default, should not show default annotation
	if strings.Contains(got, "default: ``") {
		t.Errorf("empty default should not be shown, got: %q", got)
	}
}

func TestRenderCommandTable_requiredFlag(t *testing.T) {
	meta := PrimMeta{
		Name: "testprim",
		Commands: []CmdMeta{
			{
				Name:     "purge",
				Synopsis: "purge <queue>",
				Flags: []FlagMeta{
					{Name: "status", Usage: "status to purge", Default: "", Required: true},
					{Name: "older-than", Usage: "only purge older than", Default: ""},
				},
			},
		},
	}
	got := RenderCommandTable(meta)
	if !strings.Contains(got, "*(required)*") {
		t.Errorf("expected required marker for status flag, got: %q", got)
	}
	// older-than is not required, should not have the marker
	if strings.Count(got, "*(required)*") != 1 {
		t.Errorf("expected exactly one required marker, got: %q", got)
	}
}

func TestRenderCommandTable_noFlags(t *testing.T) {
	meta := PrimMeta{
		Name: "testprim",
		Commands: []CmdMeta{
			{Name: "queues", Synopsis: "queues", Flags: nil},
		},
	}
	got := RenderCommandTable(meta)
	if !strings.Contains(got, "| — |") {
		t.Errorf("expected em dash for empty flags, got: %q", got)
	}
}

func TestRenderCommandTable_trivialDefaults(t *testing.T) {
	flags := []FlagMeta{
		{Name: "dead", Usage: "force dead", Default: "false"},
		{Name: "count", Usage: "count items", Default: "0"},
		{Name: "name", Usage: "a name", Default: ""},
	}
	got := renderFlags(flags)
	if strings.Contains(got, "default:") {
		t.Errorf("trivial defaults (false, 0, empty) should not appear, got: %q", got)
	}
}

func TestRenderFlags_empty(t *testing.T) {
	got := renderFlags(nil)
	if got != "—" {
		t.Errorf("expected em dash for nil flags, got: %q", got)
	}
}

func TestIsTrivialDefault(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"", true},
		{"false", true},
		{"0", true},
		{"normal", false},
		{"30m", false},
		{"true", false},
		{"1", false},
	}
	for _, c := range cases {
		got := isTrivialDefault(c.input)
		if got != c.want {
			t.Errorf("isTrivialDefault(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}
