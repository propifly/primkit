package cli

import (
	"fmt"
	"strings"

	"github.com/propifly/primkit/taskprim/internal/model"
	"github.com/spf13/cobra"
)

func newEditCmd() *cobra.Command {
	var (
		what      string
		list      string
		waitingOn string
		taskCtx   string
		parentID  string
		addLabels []string
		delLabels []string
	)

	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a task's fields",
		Long: `Updates one or more fields on a task. Only the specified fields are changed.

Labels can be added or removed:
  taskprim edit t_abc --add-label urgent --del-label maybe

To clear waiting_on, use an empty string:
  taskprim edit t_abc --waiting-on ""`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			id := args[0]

			update := &model.TaskUpdate{}
			changed := false

			if cmd.Flags().Changed("what") {
				update.What = &what
				changed = true
			}
			if cmd.Flags().Changed("list") {
				update.List = &list
				changed = true
			}
			if cmd.Flags().Changed("waiting-on") {
				// An empty string means "clear waiting_on".
				update.WaitingOn = &waitingOn
				changed = true
			}
			if cmd.Flags().Changed("context") {
				update.Context = &taskCtx
				changed = true
			}
			if cmd.Flags().Changed("parent") {
				update.ParentID = &parentID
				changed = true
			}
			if len(addLabels) > 0 {
				update.AddLabels = splitCSV(addLabels)
				changed = true
			}
			if len(delLabels) > 0 {
				update.DelLabels = splitCSV(delLabels)
				changed = true
			}

			if !changed {
				return fmt.Errorf("no fields specified — use flags like --what, --list, --add-label")
			}

			if err := s.UpdateTask(cmd.Context(), id, update); err != nil {
				return fmt.Errorf("editing %s: %w", id, err)
			}

			// Show the updated task.
			task, err := s.GetTask(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("fetching updated task: %w", err)
			}

			return outputTask(cmd, task)
		},
	}

	cmd.Flags().StringVar(&what, "what", "", "update the task description")
	cmd.Flags().StringVarP(&list, "list", "l", "", "move to a different list")
	cmd.Flags().StringVar(&waitingOn, "waiting-on", "", "set or clear (empty string) waiting_on")
	cmd.Flags().StringVar(&taskCtx, "context", "", "update context notes")
	cmd.Flags().StringVar(&parentID, "parent", "", "set or clear parent task ID")
	cmd.Flags().StringSliceVar(&addLabels, "add-label", nil, "add labels (repeatable)")
	cmd.Flags().StringSliceVar(&delLabels, "del-label", nil, "remove labels (repeatable)")

	return cmd
}

// splitCSV flattens comma-separated values in a string slice. This handles
// both --add-label a,b and --add-label a --add-label b.
func splitCSV(values []string) []string {
	var result []string
	for _, v := range values {
		for _, part := range strings.Split(v, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				result = append(result, part)
			}
		}
	}
	return result
}
