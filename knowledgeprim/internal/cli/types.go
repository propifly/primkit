package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newTypesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "types",
		Short: "List all entity types with counts",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			types, err := s.ListTypes(cmd.Context())
			if err != nil {
				return fmt.Errorf("listing types: %w", err)
			}

			return outputTypes(cmd, types)
		},
	}
}
