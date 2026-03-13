package cli

import "testing"

func TestGolden_Set(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{name: "text", format: "text"},
		{name: "json", format: "json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore(t)
			out := execCmd(t, s, "set", "app", "theme", `"dark"`, "-f", tt.format)
			assertGolden(t, "set_"+tt.name, out)
		})
	}
}

func TestGolden_Get(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{name: "text", format: "text"},
		{name: "json", format: "json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore(t)
			execCmd(t, s, "set", "app", "theme", `"dark"`)
			out := execCmd(t, s, "get", "app", "theme", "-f", tt.format)
			assertGolden(t, "get_"+tt.name, out)
		})
	}
}

func TestGolden_Query(t *testing.T) {
	tests := []struct {
		name   string
		format string
		seed   bool
	}{
		{name: "empty", format: "text", seed: false},
		{name: "text", format: "text", seed: true},
		{name: "json", format: "json", seed: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore(t)
			if tt.seed {
				execCmd(t, s, "set", "app", "theme", `"dark"`)
				execCmd(t, s, "set", "app", "lang", `"en"`)
			}
			out := execCmd(t, s, "query", "app", "-f", tt.format)
			assertGolden(t, "query_"+tt.name, out)
		})
	}
}

func TestGolden_Namespaces(t *testing.T) {
	tests := []struct {
		name   string
		format string
		seed   bool
	}{
		{name: "empty", format: "text", seed: false},
		{name: "text", format: "text", seed: true},
		{name: "json", format: "json", seed: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore(t)
			if tt.seed {
				execCmd(t, s, "set", "app", "theme", `"dark"`)
				execCmd(t, s, "set", "app", "lang", `"en"`)
			}
			out := execCmd(t, s, "namespaces", "-f", tt.format)
			assertGolden(t, "namespaces_"+tt.name, out)
		})
	}
}

func TestGolden_Stats(t *testing.T) {
	tests := []struct {
		name   string
		format string
		seed   bool
	}{
		{name: "text", format: "text", seed: true},
		{name: "json", format: "json", seed: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore(t)
			if tt.seed {
				execCmd(t, s, "set", "app", "theme", `"dark"`)
				execCmd(t, s, "set", "app", "lang", `"en"`)
			}
			out := execCmd(t, s, "stats", "-f", tt.format)
			assertGolden(t, "stats_"+tt.name, out)
		})
	}
}
