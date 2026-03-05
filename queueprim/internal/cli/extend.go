package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func newExtendCmd() *cobra.Command {
	var by string

	cmd := &cobra.Command{
		Use:   "extend <id>",
		Short: "Extend a claimed job's visibility timeout",
		Long: `Extends the visibility timeout of a claimed job to prevent auto-release.
Use this before starting a long-running operation on a claimed job.

  queueprim extend q_a7x3k9m2p4qr --by 45m`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			id := args[0]

			d, err := time.ParseDuration(by)
			if err != nil {
				return fmt.Errorf("invalid duration %q: %w", by, err)
			}

			if err := s.ExtendJob(cmd.Context(), id, d); err != nil {
				return fmt.Errorf("extending job: %w", err)
			}

			job, err := s.GetJob(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("fetching job: %w", err)
			}

			return outputJob(cmd, job)
		},
	}

	cmd.Flags().StringVar(&by, "by", "30m", "extension duration, e.g. 30m, 1h")

	return cmd
}
