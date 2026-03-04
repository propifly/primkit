package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDisconnectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disconnect <source-id> <target-id> <relationship>",
		Short: "Remove an edge between two entities",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			if err := s.DeleteEdge(cmd.Context(), args[0], args[1], args[2]); err != nil {
				return fmt.Errorf("deleting edge: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Disconnected %s → %s [%s]\n", args[0], args[1], args[2])
			return nil
		},
	}
}
