package cli

import (
	"errors"
	"fmt"

	"github.com/propifly/primkit/queueprim/internal/store"
	"github.com/spf13/cobra"
)

func newPeekCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "peek <queue>",
		Short: "Show the next job without claiming it",
		Long: `Shows the next job that would be returned by dequeue, without claiming it.
Useful for inspecting queue state without consuming a job.

  queueprim peek infra/fixes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			queue := args[0]

			job, err := s.PeekJob(cmd.Context(), queue)
			if err != nil {
				if errors.Is(err, store.ErrEmpty) {
					fmt.Fprintln(cmd.OutOrStdout(), "queue is empty")
					return nil
				}
				return fmt.Errorf("peeking queue: %w", err)
			}

			return outputJob(cmd, job)
		},
	}
}
