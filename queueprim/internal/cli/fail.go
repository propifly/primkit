package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newFailCmd() *cobra.Command {
	var (
		reason string
		dead   bool
	)

	cmd := &cobra.Command{
		Use:   "fail <id>",
		Short: "Mark a claimed job as failed",
		Long: `Marks a claimed job as failed. If retries remain and --dead is not set,
the job returns to pending for another attempt. Otherwise it moves to dead.

  queueprim fail q_a7x3k9m2p4qr --reason "fix applied but verification failed"
  queueprim fail q_a7x3k9m2p4qr --reason "escalate" --dead`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			id := args[0]

			if err := s.FailJob(cmd.Context(), id, reason, dead); err != nil {
				return fmt.Errorf("failing job: %w", err)
			}

			job, err := s.GetJob(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("fetching job: %w", err)
			}

			return outputJob(cmd, job)
		},
	}

	cmd.Flags().StringVar(&reason, "reason", "", "human-readable failure reason")
	cmd.Flags().BoolVar(&dead, "dead", false, "force to dead-letter regardless of retry count")

	return cmd
}
