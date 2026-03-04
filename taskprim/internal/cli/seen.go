package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSeenCmd() *cobra.Command {
	var list string

	cmd := &cobra.Command{
		Use:   "seen <agent> [task_ids...]",
		Short: "Mark tasks as seen by an agent",
		Long: `Records that an agent has seen tasks. This is used for cross-session
awareness — agents can query for tasks they haven't seen yet.

Mark specific tasks:
  taskprim seen johanna t_abc t_def

Mark all open tasks in a list:
  taskprim seen johanna --list andres`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			agent := args[0]
			taskIDs := args[1:]

			if list != "" {
				// Mark all open tasks in the list as seen.
				if err := s.MarkAllSeen(cmd.Context(), agent, list); err != nil {
					return fmt.Errorf("marking all seen: %w", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Marked all open tasks in %q as seen by %s\n", list, agent)
			} else if len(taskIDs) > 0 {
				// Mark specific tasks as seen.
				if err := s.MarkSeen(cmd.Context(), agent, taskIDs); err != nil {
					return fmt.Errorf("marking seen: %w", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Marked %d task(s) as seen by %s\n", len(taskIDs), agent)
			} else {
				return fmt.Errorf("specify task IDs or --list to mark as seen")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&list, "list", "l", "", "mark all open tasks in this list as seen")

	return cmd
}
