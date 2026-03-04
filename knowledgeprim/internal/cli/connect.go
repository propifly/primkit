package cli

import (
	"fmt"

	"github.com/propifly/primkit/knowledgeprim/internal/model"
	"github.com/spf13/cobra"
)

func newConnectCmd() *cobra.Command {
	var (
		relationship string
		edgeContext  string
		weight       float64
	)

	cmd := &cobra.Command{
		Use:   "connect <source-id> <target-id>",
		Short: "Create an edge between two entities",
		Long: `Creates a directed edge from the source entity to the target entity.

  knowledgeprim connect e_x7k2m9 e_p3n5q8 --relationship contradicts \
    --context "Both address on-device inference but reach opposite conclusions."
  knowledgeprim connect e_x7k2m9 e_p3n5q8 --relationship extends`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			edge := &model.Edge{
				SourceID:     args[0],
				TargetID:     args[1],
				Relationship: relationship,
				Weight:       weight,
			}
			if edgeContext != "" {
				edge.Context = &edgeContext
			}

			if err := s.CreateEdge(cmd.Context(), edge); err != nil {
				return fmt.Errorf("creating edge: %w", err)
			}

			w := cmd.OutOrStdout()
			if getFormat(cmd) == "json" {
				return writeJSON(w, edge)
			}
			fmt.Fprintf(w, "Connected %s → %s [%s] (weight: %.1f)\n",
				args[0], args[1], relationship, edge.Weight)
			return nil
		},
	}

	cmd.Flags().StringVar(&relationship, "relationship", "", "relationship type")
	cmd.Flags().StringVar(&edgeContext, "context", "", "edge context (why this connection exists)")
	cmd.Flags().Float64Var(&weight, "weight", 1.0, "edge weight")

	_ = cmd.MarkFlagRequired("relationship")

	return cmd
}
