package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/propifly/primkit/knowledgeprim/internal/model"
	"github.com/spf13/cobra"
)

func newImportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import",
		Short: "Import entities and edges from JSON",
		Long: `Imports entities and edges from JSON via stdin.

  knowledgeprim import < knowledge.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			var data model.ExportData
			if err := json.NewDecoder(os.Stdin).Decode(&data); err != nil {
				return fmt.Errorf("parsing import data: %w", err)
			}

			if err := s.ImportEntities(cmd.Context(), &data); err != nil {
				return fmt.Errorf("importing: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Imported %d entities and %d edges.\n",
				len(data.Entities), len(data.Edges))
			return nil
		},
	}
}
