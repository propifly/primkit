package cli

import (
	"fmt"
	"strings"

	"github.com/propifly/primkit/taskprim/internal/model"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var (
		list      string
		source    string
		waitingOn string
		parentID  string
		taskCtx   string
		labels    []string
	)

	cmd := &cobra.Command{
		Use:   "add <what>",
		Short: "Create a new task",
		Long: `Creates a task with the given description.

The task is added to the default list (from TASKPRIM_LIST env var, or "default")
unless --list is specified.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			what := strings.Join(args, " ")

			task := &model.Task{
				List:   list,
				What:   what,
				Source: source,
				Labels: labels,
			}
			if waitingOn != "" {
				task.WaitingOn = &waitingOn
			}
			if parentID != "" {
				task.ParentID = &parentID
			}
			if taskCtx != "" {
				task.Context = &taskCtx
			}

			if err := s.CreateTask(cmd.Context(), task); err != nil {
				return fmt.Errorf("creating task: %w", err)
			}

			return outputTask(cmd, task)
		},
	}

	cmd.Flags().StringVarP(&list, "list", "l", defaultList(), "list to add the task to")
	cmd.Flags().StringVarP(&source, "source", "s", "cli", "who created this task")
	cmd.Flags().StringVar(&waitingOn, "waiting-on", "", "what this task is blocked on")
	cmd.Flags().StringVar(&parentID, "parent", "", "parent task ID for subtasks")
	cmd.Flags().StringVar(&taskCtx, "context", "", "additional context or notes")
	cmd.Flags().StringSliceVar(&labels, "label", nil, "labels (repeatable or comma-separated)")

	return cmd
}
