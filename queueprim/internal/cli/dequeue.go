package cli

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/propifly/primkit/queueprim/internal/store"
	"github.com/spf13/cobra"
)

func newDequeueCmd() *cobra.Command {
	var (
		worker      string
		timeout     string
		jobType     string
		wait        bool
		waitTimeout string
	)

	cmd := &cobra.Command{
		Use:   "dequeue <queue>",
		Short: "Atomically claim the next available job",
		Long: `Atomically claims the next pending job in the given queue and returns it.

The job is exclusively locked by the worker for the given timeout duration.
If the worker crashes without completing or failing the job, it is automatically
returned to pending after the timeout expires.

Exit codes:
  0 — job claimed and printed
  1 — queue is empty (no jobs available) or error

Usage:
  queueprim dequeue infra/fixes --worker raphael --format json
  queueprim dequeue infra/fixes --wait
  queueprim dequeue infra/fixes --wait --timeout-wait 5m`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			queue := args[0]

			if worker == "" {
				if h, err := os.Hostname(); err == nil {
					worker = h
				} else {
					worker = "cli"
				}
			}

			claimTimeout := 30 * time.Minute
			if timeout != "" {
				var err error
				claimTimeout, err = time.ParseDuration(timeout)
				if err != nil {
					return fmt.Errorf("invalid timeout %q: %w", timeout, err)
				}
			}

			var waitDeadline time.Time
			if wait && waitTimeout != "" {
				d, err := time.ParseDuration(waitTimeout)
				if err != nil {
					return fmt.Errorf("invalid timeout-wait %q: %w", waitTimeout, err)
				}
				waitDeadline = time.Now().Add(d)
			}

			for {
				job, err := s.DequeueJob(cmd.Context(), queue, worker, jobType, claimTimeout)
				if err == nil {
					return outputJob(cmd, job)
				}

				if !errors.Is(err, store.ErrEmpty) {
					return fmt.Errorf("dequeuing: %w", err)
				}

				// Queue is empty.
				if !wait {
					fmt.Fprintln(cmd.ErrOrStderr(), "queue is empty")
					os.Exit(1)
				}

				// Waiting mode: check timeout.
				if !waitDeadline.IsZero() && time.Now().After(waitDeadline) {
					fmt.Fprintln(cmd.ErrOrStderr(), "queue is empty (wait timeout)")
					os.Exit(1)
				}

				time.Sleep(2 * time.Second)
			}
		},
	}

	cmd.Flags().StringVar(&worker, "worker", "", "worker name for claimed_by tracking (default: hostname)")
	cmd.Flags().StringVar(&timeout, "timeout", "30m", "visibility timeout for claimed job")
	cmd.Flags().StringVar(&jobType, "type", "", "only claim jobs of this type")
	cmd.Flags().BoolVar(&wait, "wait", false, "block until a job appears")
	cmd.Flags().StringVar(&waitTimeout, "timeout-wait", "", "max time to wait (e.g. 5m); 0 = wait forever")

	return cmd
}
