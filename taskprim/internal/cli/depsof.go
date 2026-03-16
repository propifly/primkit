package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDepsOfCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "deps-of <task-id>",
		Short: "List what depends on a task (reverse lookup)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			tasks, err := s.Dependents(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("listing dependents: %w", err)
			}
			return outputTasks(cmd, tasks)
		},
	}
}
