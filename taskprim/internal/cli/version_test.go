package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCmd(t *testing.T) {
	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"version"})

	err := root.Execute()
	require.NoError(t, err, "version command should not require a database")
}

func TestVersionCmdOutput(t *testing.T) {
	// Set the Version variable to a known value.
	old := Version
	Version = "v1.2.3"
	defer func() { Version = old }()

	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"version"})

	require.NoError(t, root.Execute())
	assert.Contains(t, buf.String(), "taskprim v1.2.3")
}

func TestResolveVersion(t *testing.T) {
	old := Version
	defer func() { Version = old }()

	// When Version is set via ldflags, use it directly.
	Version = "v0.5.0"
	assert.Equal(t, "v0.5.0", resolveVersion())

	// When Version is empty, falls back to VCS info or "dev".
	Version = ""
	v := resolveVersion()
	assert.NotEmpty(t, v)
}
