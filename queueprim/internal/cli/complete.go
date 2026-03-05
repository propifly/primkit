package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newCompleteCmd() *cobra.Command {
	var outputStr string

	cmd := &cobra.Command{
		Use:   "complete <id>",
		Short: "Mark a claimed job as done",
		Long: `Marks a claimed job as done. Optionally attach a JSON output payload.

  queueprim complete q_a7x3k9m2p4qr
  queueprim complete q_a7x3k9m2p4qr --output '{"fix_applied":"...","commit":"abc123"}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			id := args[0]

			var output []byte
			if outputStr != "" {
				if !json.Valid([]byte(outputStr)) {
					return fmt.Errorf("output must be valid JSON")
				}
				output = []byte(outputStr)
			}

			if err := s.CompleteJob(cmd.Context(), id, output); err != nil {
				return fmt.Errorf("completing job: %w", err)
			}

			job, err := s.GetJob(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("fetching job: %w", err)
			}

			return outputJob(cmd, job)
		},
	}

	cmd.Flags().StringVar(&outputStr, "output", "", "JSON output payload from the worker")

	return cmd
}
