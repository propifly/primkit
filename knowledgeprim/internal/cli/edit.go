package cli

import (
	"encoding/json"
	"fmt"

	"github.com/propifly/primkit/knowledgeprim/internal/model"
	"github.com/spf13/cobra"
)

func newEditCmd() *cobra.Command {
	var (
		title      string
		body       string
		properties string
	)

	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Update entity fields",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			update := &model.EntityUpdate{}
			if cmd.Flags().Changed("title") {
				update.Title = &title
			}
			if cmd.Flags().Changed("body") {
				update.Body = &body
			}
			if cmd.Flags().Changed("properties") {
				raw := json.RawMessage(properties)
				update.Properties = &raw
			}

			if err := s.UpdateEntity(cmd.Context(), args[0], update); err != nil {
				return fmt.Errorf("updating entity: %w", err)
			}

			// Return the updated entity.
			entity, err := s.GetEntity(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("fetching updated entity: %w", err)
			}

			return outputEntity(cmd, entity)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "new title")
	cmd.Flags().StringVar(&body, "body", "", "new body")
	cmd.Flags().StringVar(&properties, "properties", "", "JSON properties")

	return cmd
}
