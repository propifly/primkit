package cli

import (
	"fmt"

	"github.com/propifly/primkit/knowledgeprim/internal/model"
	"github.com/spf13/cobra"
)

func newEdgeEditCmd() *cobra.Command {
	var (
		edgeContext string
		weight      float64
	)

	cmd := &cobra.Command{
		Use:   "edge-edit <source-id> <target-id> <relationship>",
		Short: "Update an edge's context or weight",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			update := &model.EdgeUpdate{}
			if cmd.Flags().Changed("context") {
				update.Context = &edgeContext
			}
			if cmd.Flags().Changed("weight") {
				update.Weight = &weight
			}

			if err := s.UpdateEdge(cmd.Context(), args[0], args[1], args[2], update); err != nil {
				return fmt.Errorf("updating edge: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Updated edge %s → %s [%s]\n", args[0], args[1], args[2])
			return nil
		},
	}

	cmd.Flags().StringVar(&edgeContext, "context", "", "edge context")
	cmd.Flags().Float64Var(&weight, "weight", 0, "edge weight")

	return cmd
}
