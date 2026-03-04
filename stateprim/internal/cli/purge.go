package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPurgeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "purge <namespace> <duration>",
		Short: "Delete records older than a duration",
		Long: `Deletes records in a namespace whose updated_at is older than the given
duration. Supports Go durations (24h, 1h30m) and day notation (7d, 30d).

  stateprim purge audit 30d
  stateprim purge events 24h`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			deleted, err := s.Purge(cmd.Context(), args[0], args[1])
			if err != nil {
				return fmt.Errorf("purging records: %w", err)
			}

			switch getFormat(cmd) {
			case "json":
				return writeJSON(cmd.OutOrStdout(), map[string]int{"deleted": deleted})
			default:
				fmt.Fprintf(cmd.OutOrStdout(), "Purged %d record(s)\n", deleted)
				return nil
			}
		},
	}
}
