package cli

import (
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags:
//
//	-ldflags "-X github.com/propifly/primkit/taskprim/internal/cli.Version=v0.4.1"
//
// When not set (e.g. "go run"), it falls back to VCS info from the Go module.
var Version = ""

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the taskprim version",
		// Override parent's PersistentPreRunE — no database needed.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("taskprim " + resolveVersion())
		},
	}
}

func resolveVersion() string {
	if Version != "" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		// go install sets Main.Version to the module version.
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" && len(s.Value) >= 7 {
				return s.Value[:7]
			}
		}
	}
	return "dev"
}
