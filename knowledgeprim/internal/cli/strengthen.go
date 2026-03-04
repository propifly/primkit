package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStrengthenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "strengthen <source-id> <target-id> <relationship>",
		Short: "Increment an edge's weight",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			if err := s.StrengthenEdge(cmd.Context(), args[0], args[1], args[2]); err != nil {
				return fmt.Errorf("strengthening edge: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Strengthened %s → %s [%s]\n", args[0], args[1], args[2])
			return nil
		},
	}
}
