package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/propifly/primkit/taskprim/internal/model"
	"github.com/spf13/cobra"
)

func newImportCmd() *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import tasks from JSON",
		Long: `Imports tasks from a JSON file or stdin. The JSON should be an array of
task objects (as produced by 'taskprim export').

  taskprim import --file backup.json
  cat backup.json | taskprim import`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			// Read from file or stdin.
			var input *os.File
			if file != "" {
				f, err := os.Open(file)
				if err != nil {
					return fmt.Errorf("opening file: %w", err)
				}
				defer f.Close()
				input = f
			} else {
				input = os.Stdin
			}

			var tasks []*model.Task
			if err := json.NewDecoder(input).Decode(&tasks); err != nil {
				return fmt.Errorf("parsing JSON: %w", err)
			}

			if err := s.ImportTasks(cmd.Context(), tasks); err != nil {
				return fmt.Errorf("importing tasks: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Imported %d task(s)\n", len(tasks))
			return nil
		},
	}

	cmd.Flags().StringVarP(&file, "file", "F", "", "path to JSON file (default: stdin)")

	return cmd
}
