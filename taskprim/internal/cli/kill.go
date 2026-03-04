package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newKillCmd() *cobra.Command {
	var reason string

	cmd := &cobra.Command{
		Use:   "kill <id>",
		Short: "Kill a task (drop it with a reason)",
		Long: `Marks a task as killed. A reason is required — this ensures nothing falls
through the cracks silently. If a task is no longer relevant, say why.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			if reason == "" {
				return fmt.Errorf("--reason is required when killing a task")
			}

			if err := s.KillTask(cmd.Context(), args[0], reason); err != nil {
				return fmt.Errorf("killing %s: %w", args[0], err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Killed: %s (%s)\n", args[0], reason)
			return nil
		},
	}

	cmd.Flags().StringVarP(&reason, "reason", "r", "", "why this task is being dropped (required)")

	return cmd
}
