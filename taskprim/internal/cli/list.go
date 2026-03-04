package cli

import (
	"fmt"
	"time"

	"github.com/propifly/primkit/taskprim/internal/model"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var (
		list     string
		state    string
		labels   []string
		source   string
		waiting  bool
		unseenBy string
		seenBy   string
		since    string
		stale    string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Long: `Lists tasks matching the given filters. Without filters, shows all tasks.

Examples:
  taskprim list --list andres --label today
  taskprim list --unseen-by johanna
  taskprim list --stale 7d`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			filter := &model.Filter{}
			if list != "" {
				filter.List = list
			}
			if state != "" {
				st := model.State(state)
				filter.State = &st
			}
			if len(labels) > 0 {
				filter.Labels = labels
			}
			if source != "" {
				filter.Source = source
			}
			if cmd.Flags().Changed("waiting") {
				filter.Waiting = &waiting
			}
			if unseenBy != "" {
				filter.UnseenBy = unseenBy
			}
			if seenBy != "" {
				filter.SeenBy = seenBy
				if since != "" {
					d, err := parseDuration(since)
					if err != nil {
						return fmt.Errorf("invalid --since value: %w", err)
					}
					filter.Since = d
				}
			}
			if stale != "" {
				d, err := parseDuration(stale)
				if err != nil {
					return fmt.Errorf("invalid --stale value: %w", err)
				}
				filter.Stale = d
			}

			tasks, err := s.ListTasks(cmd.Context(), filter)
			if err != nil {
				return fmt.Errorf("listing tasks: %w", err)
			}

			return outputTasks(cmd, tasks)
		},
	}

	cmd.Flags().StringVarP(&list, "list", "l", "", "filter by list")
	cmd.Flags().StringVar(&state, "state", "", "filter by state: open, done, killed")
	cmd.Flags().StringSliceVar(&labels, "label", nil, "filter by label (repeatable, AND logic)")
	cmd.Flags().StringVar(&source, "source", "", "filter by source")
	cmd.Flags().BoolVar(&waiting, "waiting", false, "only tasks with waiting_on set")
	cmd.Flags().StringVar(&unseenBy, "unseen-by", "", "tasks not seen by this agent")
	cmd.Flags().StringVar(&seenBy, "seen-by", "", "tasks seen by this agent (use with --since)")
	cmd.Flags().StringVar(&since, "since", "", "time window for --seen-by (e.g., 24h, 7d)")
	cmd.Flags().StringVar(&stale, "stale", "", "tasks not updated within duration (e.g., 7d)")

	return cmd
}

// parseDuration handles both Go-style durations (24h, 30m) and short-form
// day notation (7d, 30d) which Go's time.ParseDuration doesn't support.
func parseDuration(s string) (time.Duration, error) {
	// Handle "Nd" notation by converting to hours.
	if len(s) > 1 && s[len(s)-1] == 'd' {
		var days int
		if _, err := fmt.Sscanf(s, "%dd", &days); err == nil {
			return time.Duration(days) * 24 * time.Hour, nil
		}
	}
	return time.ParseDuration(s)
}
