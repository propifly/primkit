package cli

import (
	"encoding/json"
	"fmt"

	"github.com/propifly/primkit/stateprim/internal/model"
	"github.com/spf13/cobra"
)

func newSetCmd() *cobra.Command {
	var immutable bool

	cmd := &cobra.Command{
		Use:   "set <namespace> <key> <json-value>",
		Short: "Create or update a record",
		Long: `Creates or updates a record (upsert). The value must be valid JSON.

  stateprim set config theme '"dark"'
  stateprim set agents/johanna last_run '{"at":"2026-03-03T10:00:00Z"}'`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			value := json.RawMessage(args[2])
			if !json.Valid(value) {
				return fmt.Errorf("value must be valid JSON")
			}

			r := &model.Record{
				Namespace: args[0],
				Key:       args[1],
				Value:     value,
				Immutable: immutable,
			}

			if err := s.Set(cmd.Context(), r); err != nil {
				return fmt.Errorf("setting record: %w", err)
			}

			return outputRecord(cmd, r)
		},
	}

	cmd.Flags().BoolVar(&immutable, "immutable", false, "mark the record as immutable")

	return cmd
}
