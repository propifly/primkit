package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/propifly/primkit/queueprim/internal/model"
	"github.com/spf13/cobra"
)

func newImportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import",
		Short: "Import jobs from JSON (stdin)",
		Long: `Imports jobs from a JSON file on stdin. Preserves original IDs.

  queueprim import < backup.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			var jobs []*model.Job
			if err := json.NewDecoder(os.Stdin).Decode(&jobs); err != nil {
				return fmt.Errorf("decoding input: %w", err)
			}

			if err := s.ImportJobs(cmd.Context(), jobs); err != nil {
				return fmt.Errorf("importing jobs: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Imported %d job(s).\n", len(jobs))
			return nil
		},
	}
}
