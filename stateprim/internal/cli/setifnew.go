package cli

import (
	"encoding/json"
	"fmt"

	"github.com/propifly/primkit/stateprim/internal/model"
	"github.com/spf13/cobra"
)

func newSetIfNewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-if-new <namespace> <key> <json-value>",
		Short: "Create a record only if the key doesn't exist",
		Long: `Atomically creates a record if the key doesn't already exist. Returns an
error if the key is already present. Useful for dedup operations.

  stateprim set-if-new sent-emails msg:abc123 '{"to":"alice@example.com"}'`,
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
			}

			if err := s.SetIfNew(cmd.Context(), r); err != nil {
				return fmt.Errorf("set-if-new: %w", err)
			}

			return outputRecord(cmd, r)
		},
	}
}
