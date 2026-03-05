package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show aggregate job statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			stats, err := s.Stats(cmd.Context())
			if err != nil {
				return fmt.Errorf("getting stats: %w", err)
			}

			return outputStats(cmd, stats)
		},
	}
}
