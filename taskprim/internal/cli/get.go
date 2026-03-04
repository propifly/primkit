package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Show task details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			task, err := s.GetTask(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("getting task: %w", err)
			}

			return outputTask(cmd, task)
		},
	}
}
