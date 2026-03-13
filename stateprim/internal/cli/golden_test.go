package cli

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

var (
	reID        = regexp.MustCompile(`[teqas]_[a-z0-9]{12}`)
	reTimestamp = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z`)
)

func normalizeOutput(s string) string {
	s = reID.ReplaceAllString(s, "<ID>")
	s = reTimestamp.ReplaceAllString(s, "<TIMESTAMP>")
	return s
}

func assertGolden(t *testing.T, name, got string) {
	t.Helper()
	golden := filepath.Join("testdata", name+".golden")
	got = normalizeOutput(got)

	if *update {
		require.NoError(t, os.MkdirAll("testdata", 0o755))
		require.NoError(t, os.WriteFile(golden, []byte(got), 0o644))
		return
	}

	expected, err := os.ReadFile(golden)
	require.NoError(t, err, "golden file %s not found — run with -update", golden)
	assert.Equal(t, string(expected), got)
}
