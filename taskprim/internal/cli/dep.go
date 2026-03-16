package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDepCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dep",
		Short: "Manage task dependencies",
		Long: `Manage structural task-to-task dependencies.

A dependency edge "A depends on B" means A cannot start until B is resolved.
These are distinct from waiting_on (freeform text for external blockers).

Examples:
  taskprim dep add <task-id> <depends-on-id>
  taskprim dep rm  <task-id> <depends-on-id>
  taskprim dep ls  <task-id>`,
	}

	cmd.AddCommand(newDepAddCmd(), newDepRmCmd(), newDepLsCmd())
	return cmd
}

func newDepAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <task-id> <depends-on-id>",
		Short: "Add a dependency edge",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			if err := s.AddDep(cmd.Context(), args[0], args[1]); err != nil {
				return fmt.Errorf("adding dep: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Added: %s depends on %s\n", args[0], args[1])
			return nil
		},
	}
}

func newDepRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <task-id> <depends-on-id>",
		Short: "Remove a dependency edge",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			if err := s.RemoveDep(cmd.Context(), args[0], args[1]); err != nil {
				return fmt.Errorf("removing dep: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed: %s no longer depends on %s\n", args[0], args[1])
			return nil
		},
	}
}

func newDepLsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls <task-id>",
		Short: "List what a task depends on",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			tasks, err := s.Deps(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("listing deps: %w", err)
			}
			return outputTasks(cmd, tasks)
		},
	}
}
