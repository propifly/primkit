package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a specific job by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			id := args[0]

			job, err := s.GetJob(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("getting job: %w", err)
			}

			return outputJob(cmd, job)
		},
	}
}
