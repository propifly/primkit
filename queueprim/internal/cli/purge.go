package cli

import (
	"fmt"
	"time"

	"github.com/propifly/primkit/queueprim/internal/model"
	"github.com/spf13/cobra"
)

func newPurgeCmd() *cobra.Command {
	var (
		status    string
		olderThan string
	)

	cmd := &cobra.Command{
		Use:   "purge <queue>",
		Short: "Delete jobs from a queue by status",
		Long: `Deletes jobs from a queue matching the given status and optional age filter.

  queueprim purge infra/fixes --status done --older-than 7d
  queueprim purge infra/fixes --status dead`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			queue := args[0]

			if status == "" {
				return fmt.Errorf("--status is required")
			}

			var olderThanDur time.Duration
			if olderThan != "" {
				var err error
				olderThanDur, err = parseDurationFlag(olderThan)
				if err != nil {
					return fmt.Errorf("invalid older-than: %w", err)
				}
			}

			n, err := s.PurgeJobs(cmd.Context(), queue, model.Status(status), olderThanDur)
			if err != nil {
				return fmt.Errorf("purging jobs: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted %d job(s).\n", n)
			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "status to purge: done, dead, failed (required)")
	cmd.Flags().StringVar(&olderThan, "older-than", "", "only purge jobs older than this duration, e.g. 7d, 24h")
	_ = cmd.MarkFlagRequired("status")

	return cmd
}
