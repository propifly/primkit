package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newHasCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "has <namespace> <key>",
		Short: "Check if a key exists",
		Long: `Checks whether a key exists in a namespace. Useful for dedup logic.

  stateprim has sent-emails msg:abc123

Exits 0 if the key exists, 1 if it does not.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			exists, err := s.Has(cmd.Context(), args[0], args[1])
			if err != nil {
				return fmt.Errorf("checking existence: %w", err)
			}

			switch getFormat(cmd) {
			case "json":
				return writeJSON(cmd.OutOrStdout(), map[string]bool{"exists": exists})
			case "quiet":
				if !exists {
					return fmt.Errorf("not found")
				}
				return nil
			default:
				if exists {
					fmt.Fprintln(cmd.OutOrStdout(), "yes")
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "no")
				}
				return nil
			}
		},
	}
}
