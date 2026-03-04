package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newAppendCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "append <namespace> <json-value>",
		Short: "Append an immutable record with an auto-generated key",
		Long: `Creates an immutable record with a timestamp-based key. Append records
cannot be updated or overwritten.

  stateprim append audit '{"action":"login","user":"alice"}'
  stateprim append events '{"type":"deploy","version":"1.2.3"}'`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			value := []byte(args[1])
			if !json.Valid(value) {
				return fmt.Errorf("value must be valid JSON")
			}

			r, err := s.Append(cmd.Context(), args[0], value)
			if err != nil {
				return fmt.Errorf("appending record: %w", err)
			}

			return outputRecord(cmd, r)
		},
	}
}
