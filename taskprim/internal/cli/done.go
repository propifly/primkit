package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDoneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "done <id> [id...]",
		Short: "Mark tasks as done",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			for _, id := range args {
				if err := s.DoneTask(cmd.Context(), id); err != nil {
					return fmt.Errorf("completing %s: %w", id, err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Done: %s\n", id)
			}

			return nil
		},
	}
}
