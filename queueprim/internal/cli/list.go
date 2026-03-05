package cli

import (
	"fmt"

	"github.com/propifly/primkit/queueprim/internal/model"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var (
		queue     string
		status    string
		jobType   string
		olderThan string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List jobs with optional filters",
		Long: `Lists jobs with optional filters. An empty filter returns all jobs.

  queueprim list --queue infra/fixes
  queueprim list --status dead
  queueprim list --queue infra/fixes --status claimed --older-than 1h
  queueprim list --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			filter := &model.Filter{
				Queue: queue,
				Type:  jobType,
			}

			if status != "" {
				st := model.Status(status)
				filter.Status = &st
			}

			if olderThan != "" {
				d, err := parseDurationFlag(olderThan)
				if err != nil {
					return fmt.Errorf("invalid older-than %q: %w", olderThan, err)
				}
				filter.OlderThan = d
			}

			jobs, err := s.ListJobs(cmd.Context(), filter)
			if err != nil {
				return fmt.Errorf("listing jobs: %w", err)
			}

			return outputJobs(cmd, jobs)
		},
	}

	cmd.Flags().StringVar(&queue, "queue", "", "filter to this queue")
	cmd.Flags().StringVar(&status, "status", "", "filter by status: pending, claimed, done, failed, dead")
	cmd.Flags().StringVar(&jobType, "type", "", "filter by job type")
	cmd.Flags().StringVar(&olderThan, "older-than", "", "only jobs created before now-duration, e.g. 1h")

	return cmd
}
