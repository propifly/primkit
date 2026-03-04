package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newListsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lists",
		Short: "Show all lists with task counts",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			lists, err := s.ListLists(cmd.Context())
			if err != nil {
				return fmt.Errorf("listing lists: %w", err)
			}

			return outputLists(cmd, lists)
		},
	}
}
