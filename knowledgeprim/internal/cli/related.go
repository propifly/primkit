package cli

import (
	"fmt"

	"github.com/propifly/primkit/knowledgeprim/internal/model"
	"github.com/spf13/cobra"
)

func newRelatedCmd() *cobra.Command {
	var (
		depth        int
		relationship string
		direction    string
		minWeight    float64
	)

	cmd := &cobra.Command{
		Use:   "related <id>",
		Short: "Traverse the graph from an entity",
		Long: `Shows entities related to the given entity via graph traversal.

  knowledgeprim related e_x7k2m9
  knowledgeprim related e_x7k2m9 --depth 3
  knowledgeprim related e_x7k2m9 --relationship contradicts`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			opts := &model.TraversalOpts{
				Depth:        depth,
				Relationship: relationship,
				Direction:    direction,
				MinWeight:    minWeight,
			}

			results, err := s.Related(cmd.Context(), args[0], opts)
			if err != nil {
				return fmt.Errorf("traversing graph: %w", err)
			}

			return outputTraversalResults(cmd, results)
		},
	}

	cmd.Flags().IntVar(&depth, "depth", 1, "traversal depth (hops)")
	cmd.Flags().StringVar(&relationship, "relationship", "", "filter by relationship type")
	cmd.Flags().StringVar(&direction, "direction", "both", "edge direction: outgoing, incoming, both")
	cmd.Flags().Float64Var(&minWeight, "min-weight", 0, "minimum edge weight")

	return cmd
}
