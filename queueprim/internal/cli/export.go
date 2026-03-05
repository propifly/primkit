package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	var queue string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export jobs as JSON to stdout",
		Long: `Exports all jobs (or jobs from a specific queue) as JSON to stdout.

  queueprim export > backup.json
  queueprim export --queue infra/fixes > infra-fixes.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			jobs, err := s.ExportJobs(cmd.Context(), queue)
			if err != nil {
				return fmt.Errorf("exporting jobs: %w", err)
			}

			if jobs == nil {
				jobs = nil // empty slice marshals as []
			}

			return writeJSON(cmd.OutOrStdout(), jobs)
		},
	}

	cmd.Flags().StringVar(&queue, "queue", "", "export only jobs in this queue (default: all)")

	return cmd
}
