package cli

import (
	"bytes"
	"context"
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/propifly/primkit/knowledgeprim/internal/store"
	"github.com/propifly/primkit/primkit/db"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

var (
	reID        = regexp.MustCompile(`[teqas]_[a-z0-9]{12}`)
	reTimestamp = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z`)
	reScore     = regexp.MustCompile(`\d+\.\d{4,}(?:e[+-]?\d+)?`)
	reDBSize    = regexp.MustCompile(`DB Size:\s+\d+(\.\d+)?\s*\w*B`)
)

func normalizeOutput(s string) string {
	s = reID.ReplaceAllString(s, "<ID>")
	s = reTimestamp.ReplaceAllString(s, "<TIMESTAMP>")
	s = reScore.ReplaceAllString(s, "<SCORE>")
	s = reDBSize.ReplaceAllString(s, "DB Size:   <SIZE>")
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

func goldenTestStore(t *testing.T) store.Store {
	t.Helper()
	database, err := db.OpenInMemory()
	require.NoError(t, err)
	s, err := store.NewFromDB(database)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })
	return s
}

func goldenExecCmd(t *testing.T, s store.Store, args ...string) string {
	t.Helper()
	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		ctx := context.WithValue(cmd.Context(), storeKey, s)
		ctx = context.WithValue(ctx, embedderKey, nil)
		cmd.SetContext(ctx)
		return nil
	}
	root.PersistentPostRunE = nil

	err := root.Execute()
	require.NoError(t, err)
	return buf.String()
}
