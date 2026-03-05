package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newReleaseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "release <id>",
		Short: "Return a claimed job to pending (unclaim)",
		Long: `Returns a claimed job to pending immediately, clearing the claim.
The job becomes available for dequeue again right away.

  queueprim release q_a7x3k9m2p4qr`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			id := args[0]

			if err := s.ReleaseJob(cmd.Context(), id); err != nil {
				return fmt.Errorf("releasing job: %w", err)
			}

			job, err := s.GetJob(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("fetching job: %w", err)
			}

			return outputJob(cmd, job)
		},
	}
}
