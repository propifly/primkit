package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get an entity with its edges",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			entity, err := s.GetEntity(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("getting entity: %w", err)
			}

			return outputEntity(cmd, entity)
		},
	}
}
