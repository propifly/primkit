package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <namespace> <key>",
		Short: "Retrieve a record by namespace and key",
		Long: `Retrieves a single record and prints its details.

  stateprim get config theme
  stateprim get agents/johanna last_run --format json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			r, err := s.Get(cmd.Context(), args[0], args[1])
			if err != nil {
				return fmt.Errorf("getting record: %w", err)
			}

			return outputRecord(cmd, r)
		},
	}
}
