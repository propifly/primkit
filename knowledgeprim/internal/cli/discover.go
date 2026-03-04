package cli

import (
	"fmt"

	"github.com/propifly/primkit/knowledgeprim/internal/model"
	"github.com/spf13/cobra"
)

func newDiscoverCmd() *cobra.Command {
	var (
		orphans   bool
		clusters  bool
		bridges   bool
		temporal  bool
		weakEdges bool
	)

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover patterns in the knowledge graph",
		Long: `Analyzes the graph structure to surface emergent patterns.

  knowledgeprim discover                 # full discovery report
  knowledgeprim discover --orphans       # entities with no edges
  knowledgeprim discover --clusters      # densely connected groups
  knowledgeprim discover --bridges       # cross-cluster connectors
  knowledgeprim discover --temporal      # type distribution over time
  knowledgeprim discover --weak-edges    # edges missing context`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			// If no specific flag is set, run all.
			opts := &model.DiscoverOpts{
				Orphans:   orphans,
				Clusters:  clusters,
				Bridges:   bridges,
				Temporal:  temporal,
				WeakEdges: weakEdges,
			}
			if !orphans && !clusters && !bridges && !temporal && !weakEdges {
				opts = model.DiscoverAll()
			}

			report, err := s.Discover(cmd.Context(), opts)
			if err != nil {
				return fmt.Errorf("running discovery: %w", err)
			}

			return outputDiscover(cmd, report)
		},
	}

	cmd.Flags().BoolVar(&orphans, "orphans", false, "find entities with no edges")
	cmd.Flags().BoolVar(&clusters, "clusters", false, "find densely connected groups")
	cmd.Flags().BoolVar(&bridges, "bridges", false, "find cross-cluster connectors")
	cmd.Flags().BoolVar(&temporal, "temporal", false, "show type distribution over time")
	cmd.Flags().BoolVar(&weakEdges, "weak-edges", false, "find edges with no context")

	return cmd
}
