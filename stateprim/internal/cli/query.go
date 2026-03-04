package cli

import (
	"fmt"
	"time"

	"github.com/propifly/primkit/stateprim/internal/model"
	"github.com/spf13/cobra"
)

func newQueryCmd() *cobra.Command {
	var (
		prefix    string
		since     string
		countOnly bool
	)

	cmd := &cobra.Command{
		Use:   "query <namespace>",
		Short: "Query records in a namespace",
		Long: `Lists records in a namespace with optional filters.

  stateprim query config
  stateprim query audit --prefix a_2026
  stateprim query events --since 24h
  stateprim query config --count`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			filter := &model.QueryFilter{
				Namespace: args[0],
				KeyPrefix: prefix,
				CountOnly: countOnly,
			}

			if since != "" {
				d, err := parseDuration(since)
				if err != nil {
					return fmt.Errorf("invalid --since value: %w", err)
				}
				filter.Since = d
			}

			records, count, err := s.Query(cmd.Context(), filter)
			if err != nil {
				return fmt.Errorf("querying records: %w", err)
			}

			if countOnly {
				switch getFormat(cmd) {
				case "json":
					return writeJSON(cmd.OutOrStdout(), map[string]int{"count": count})
				default:
					fmt.Fprintf(cmd.OutOrStdout(), "%d\n", count)
					return nil
				}
			}

			return outputRecords(cmd, records)
		},
	}

	cmd.Flags().StringVar(&prefix, "prefix", "", "filter by key prefix")
	cmd.Flags().StringVar(&since, "since", "", "only records updated within duration (e.g., 24h, 7d)")
	cmd.Flags().BoolVar(&countOnly, "count", false, "return count only")

	return cmd
}

// parseDuration handles both Go-style durations (24h, 30m) and short-form
// day notation (7d, 30d) which Go's time.ParseDuration doesn't support.
func parseDuration(s string) (time.Duration, error) {
	if len(s) > 1 && s[len(s)-1] == 'd' {
		var days int
		if _, err := fmt.Sscanf(s, "%dd", &days); err == nil {
			return time.Duration(days) * 24 * time.Hour, nil
		}
	}
	return time.ParseDuration(s)
}
