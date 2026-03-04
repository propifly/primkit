package cli

import (
	"fmt"

	"github.com/propifly/primkit/taskprim/internal/model"
	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	var (
		list  string
		state string
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export tasks as JSON",
		Long: `Exports tasks as a JSON array to stdout. Pipe to a file for backup:
  taskprim export > backup.json
  taskprim export --list andres --state open > active.json`,
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

			tasks, err := s.ExportTasks(cmd.Context(), filter)
			if err != nil {
				return fmt.Errorf("exporting tasks: %w", err)
			}

			return writeJSON(cmd.OutOrStdout(), tasks)
		},
	}

	cmd.Flags().StringVarP(&list, "list", "l", "", "export only tasks from this list")
	cmd.Flags().StringVar(&state, "state", "", "export only tasks in this state")

	return cmd
}
