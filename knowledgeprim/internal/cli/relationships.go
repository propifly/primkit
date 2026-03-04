package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRelationshipsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "relationships",
		Short: "List all relationship types with counts",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			rels, err := s.ListRelationships(cmd.Context())
			if err != nil {
				return fmt.Errorf("listing relationships: %w", err)
			}

			return outputRelationships(cmd, rels)
		},
	}
}
