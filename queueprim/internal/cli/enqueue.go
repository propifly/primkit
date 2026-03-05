package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/propifly/primkit/queueprim/internal/model"
	"github.com/spf13/cobra"
)

func newEnqueueCmd() *cobra.Command {
	var (
		jobType    string
		priority   string
		maxRetries int
		delay      string
	)

	cmd := &cobra.Command{
		Use:   "enqueue <queue> <json_payload>",
		Short: "Enqueue a new job",
		Long: `Enqueues a new job in the given queue with the given JSON payload.

The payload is arbitrary JSON passed as a positional argument:

  queueprim enqueue infra/fixes '{"agent":"clawson","summary":"Pi down"}' --priority high
  queueprim enqueue review/code '{"pr":42}' --type pull_request --max-retries 2
  queueprim enqueue deploy/queue '{"env":"prod"}' --delay 5m`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			queue := args[0]
			payloadStr := args[1]

			if !json.Valid([]byte(payloadStr)) {
				return fmt.Errorf("payload must be valid JSON")
			}

			job := &model.Job{
				Queue:      queue,
				Type:       jobType,
				Priority:   model.Priority(priority),
				Payload:    json.RawMessage(payloadStr),
				MaxRetries: maxRetries,
			}

			if delay != "" {
				d, err := time.ParseDuration(delay)
				if err != nil {
					return fmt.Errorf("invalid delay %q: %w", delay, err)
				}
				job.VisibleAfter = time.Now().UTC().Add(d)
			}

			if err := s.EnqueueJob(cmd.Context(), job); err != nil {
				return fmt.Errorf("enqueueing job: %w", err)
			}

			return outputJob(cmd, job)
		},
	}

	cmd.Flags().StringVar(&jobType, "type", "", "job type category for workers (e.g. ssh_auth_fail)")
	cmd.Flags().StringVar(&priority, "priority", "normal", "priority: high, normal, or low")
	cmd.Flags().IntVar(&maxRetries, "max-retries", 0, "max retries before dead-letter (default 0 = one-shot)")
	cmd.Flags().StringVar(&delay, "delay", "", "delay before job is visible, e.g. 5m, 1h")

	return cmd
}
