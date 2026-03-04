package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <namespace> <key>",
		Short: "Delete a record",
		Long: `Deletes a record by namespace and key. Returns an error if the record
does not exist.

  stateprim delete config theme`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			if err := s.Delete(cmd.Context(), args[0], args[1]); err != nil {
				return fmt.Errorf("deleting record: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Deleted.")
			return nil
		},
	}
}
