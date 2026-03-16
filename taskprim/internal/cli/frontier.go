package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newFrontierCmd() *cobra.Command {
	var list string

	cmd := &cobra.Command{
		Use:   "frontier",
		Short: "List tasks ready for execution",
		Long: `Shows open tasks where all dependencies are resolved (done or killed)
or the task has no dependencies at all. This is the "what can I work on next?" query.

Examples:
  taskprim frontier
  taskprim frontier --list work`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			tasks, err := s.Frontier(cmd.Context(), list)
			if err != nil {
				return fmt.Errorf("querying frontier: %w", err)
			}
			return outputTasks(cmd, tasks)
		},
	}

	cmd.Flags().StringVarP(&list, "list", "l", "", "filter by list")
	return cmd
}
