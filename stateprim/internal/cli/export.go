package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	var namespace string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export records as JSON",
		Long: `Exports records as a JSON array to stdout. Pipe to a file for backup:

  stateprim export > backup.json
  stateprim export --namespace config > config_backup.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			records, err := s.ExportRecords(cmd.Context(), namespace)
			if err != nil {
				return fmt.Errorf("exporting records: %w", err)
			}

			return writeJSON(cmd.OutOrStdout(), records)
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "export only this namespace")

	return cmd
}
