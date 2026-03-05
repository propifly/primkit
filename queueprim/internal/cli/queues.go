package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newQueuesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "queues",
		Short: "List all queues with job counts by status",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			queues, err := s.ListQueues(cmd.Context())
			if err != nil {
				return fmt.Errorf("listing queues: %w", err)
			}

			return outputQueues(cmd, queues)
		},
	}
}
