package cli

import (
	"fmt"

	"github.com/propifly/primkit/knowledgeprim/internal/model"
	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	var entityType string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export entities and edges as JSON",
		Long: `Exports entities and edges as JSON to stdout.

  knowledgeprim export > knowledge.json
  knowledgeprim export --type article > articles.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			filter := &model.ExportFilter{}
			if entityType != "" {
				filter.Type = entityType
			}

			data, err := s.ExportEntities(cmd.Context(), filter)
			if err != nil {
				return fmt.Errorf("exporting: %w", err)
			}

			return writeJSON(cmd.OutOrStdout(), data)
		},
	}

	cmd.Flags().StringVar(&entityType, "type", "", "export only entities of this type")

	return cmd
}
