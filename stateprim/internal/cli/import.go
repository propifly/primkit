package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/propifly/primkit/stateprim/internal/model"
	"github.com/spf13/cobra"
)

func newImportCmd() *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import records from JSON",
		Long: `Imports records from a JSON file or stdin. The JSON should be an array of
record objects (as produced by 'stateprim export').

  stateprim import --file backup.json
  cat backup.json | stateprim import`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			var input *os.File
			if file != "" {
				f, err := os.Open(file)
				if err != nil {
					return fmt.Errorf("opening file: %w", err)
				}
				defer f.Close()
				input = f
			} else {
				input = os.Stdin
			}

			var records []*model.Record
			if err := json.NewDecoder(input).Decode(&records); err != nil {
				return fmt.Errorf("parsing JSON: %w", err)
			}

			if err := s.ImportRecords(cmd.Context(), records); err != nil {
				return fmt.Errorf("importing records: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Imported %d record(s)\n", len(records))
			return nil
		},
	}

	cmd.Flags().StringVarP(&file, "file", "F", "", "path to JSON file (default: stdin)")

	return cmd
}
