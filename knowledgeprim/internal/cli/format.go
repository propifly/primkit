package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/propifly/primkit/knowledgeprim/internal/model"
	"github.com/spf13/cobra"
)

func getFormat(cmd *cobra.Command) string {
	f, _ := cmd.Flags().GetString("format")
	if f == "" {
		f, _ = cmd.InheritedFlags().GetString("format")
	}
	if f == "" {
		return "text"
	}
	return f
}

func writeJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func outputEntity(cmd *cobra.Command, entity *model.Entity) error {
	w := cmd.OutOrStdout()
	switch getFormat(cmd) {
	case "json":
		return writeJSON(w, entity)
	default:
		return writeEntityDetail(w, entity)
	}
}

func writeEntityDetail(w io.Writer, e *model.Entity) error {
	fmt.Fprintf(w, "ID:       %s\n", e.ID)
	fmt.Fprintf(w, "Type:     %s\n", e.Type)
	fmt.Fprintf(w, "Title:    %s\n", e.Title)
	if e.Body != nil {
		body := truncate(*e.Body, 200)
		fmt.Fprintf(w, "Body:     %s\n", body)
	}
	if e.URL != nil {
		fmt.Fprintf(w, "URL:      %s\n", *e.URL)
	}
	fmt.Fprintf(w, "Source:   %s\n", e.Source)
	if len(e.Properties) > 0 {
		fmt.Fprintf(w, "Props:    %s\n", string(e.Properties))
	}
	fmt.Fprintf(w, "Created:  %s\n", e.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "Updated:  %s\n", e.UpdatedAt.Format(time.RFC3339))

	if len(e.Edges) > 0 {
		fmt.Fprintf(w, "\nEdges (%d):\n", len(e.Edges))
		for _, edge := range e.Edges {
			dir := "→"
			other := edge.TargetID
			if edge.TargetID == e.ID {
				dir = "←"
				other = edge.SourceID
			}
			ctx := ""
			if edge.Context != nil && *edge.Context != "" {
				ctx = fmt.Sprintf(" [%s]", truncate(*edge.Context, 60))
			}
			fmt.Fprintf(w, "  %s %s %s (%.1f)%s\n", dir, edge.Relationship, other, edge.Weight, ctx)
		}
	}
	return nil
}

func outputSearchResults(cmd *cobra.Command, results []*model.SearchResult) error {
	w := cmd.OutOrStdout()
	switch getFormat(cmd) {
	case "json":
		return writeJSON(w, results)
	default:
		if len(results) == 0 {
			fmt.Fprintln(w, "No results found.")
			return nil
		}
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "ID\tTYPE\tSCORE\tTITLE")
		for _, r := range results {
			title := truncate(r.Entity.Title, 60)
			fmt.Fprintf(tw, "%s\t%s\t%.4f\t%s\n", r.Entity.ID, r.Entity.Type, r.Score, title)
		}
		return tw.Flush()
	}
}

func outputTraversalResults(cmd *cobra.Command, results []*model.TraversalResult) error {
	w := cmd.OutOrStdout()
	switch getFormat(cmd) {
	case "json":
		return writeJSON(w, results)
	default:
		if len(results) == 0 {
			fmt.Fprintln(w, "No related entities found.")
			return nil
		}
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "DEPTH\tDIR\tRELATIONSHIP\tWEIGHT\tID\tTITLE")
		for _, r := range results {
			dir := "→"
			if r.Direction == "incoming" {
				dir = "←"
			}
			title := truncate(r.Entity.Title, 50)
			fmt.Fprintf(tw, "%d\t%s\t%s\t%.1f\t%s\t%s\n",
				r.Depth, dir, r.Relationship, r.Weight, r.Entity.ID, title)
		}
		return tw.Flush()
	}
}

func outputDiscover(cmd *cobra.Command, report *model.DiscoverReport) error {
	w := cmd.OutOrStdout()
	switch getFormat(cmd) {
	case "json":
		return writeJSON(w, report)
	default:
		return writeDiscoverText(w, report)
	}
}

func writeDiscoverText(w io.Writer, r *model.DiscoverReport) error {
	if r.Orphans != nil {
		fmt.Fprintf(w, "=== Orphans (%d) ===\n", len(r.Orphans))
		for _, e := range r.Orphans {
			fmt.Fprintf(w, "  %s  [%s]  %s\n", e.ID, e.Type, truncate(e.Title, 60))
		}
		fmt.Fprintln(w)
	}
	if r.Clusters != nil {
		fmt.Fprintf(w, "=== Clusters (%d) ===\n", len(r.Clusters))
		for i, c := range r.Clusters {
			fmt.Fprintf(w, "  Cluster %d (%d entities):\n", i+1, c.Size)
			for _, e := range c.Entities {
				fmt.Fprintf(w, "    %s  [%s]  %s\n", e.ID, e.Type, truncate(e.Title, 50))
			}
		}
		fmt.Fprintln(w)
	}
	if r.Bridges != nil {
		fmt.Fprintf(w, "=== Bridges (%d) ===\n", len(r.Bridges))
		for _, b := range r.Bridges {
			fmt.Fprintf(w, "  %s  [%s]  %s  (%d edges)\n",
				b.Entity.ID, b.Entity.Type, truncate(b.Entity.Title, 50), b.EdgeCount)
		}
		fmt.Fprintln(w)
	}
	if r.Temporal != nil {
		fmt.Fprintf(w, "=== Temporal ===\n")
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "  PERIOD\tTYPE\tCOUNT")
		for _, g := range r.Temporal {
			fmt.Fprintf(tw, "  %s\t%s\t%d\n", g.Period, g.Type, g.Count)
		}
		tw.Flush()
		fmt.Fprintln(w)
	}
	if r.WeakEdges != nil {
		fmt.Fprintf(w, "=== Weak Edges (%d) ===\n", len(r.WeakEdges))
		for _, e := range r.WeakEdges {
			fmt.Fprintf(w, "  %s → %s [%s] (%.1f)\n",
				e.SourceID, e.TargetID, e.Relationship, e.Weight)
		}
	}
	return nil
}

func outputTypes(cmd *cobra.Command, types []model.TypeCount) error {
	w := cmd.OutOrStdout()
	switch getFormat(cmd) {
	case "json":
		return writeJSON(w, types)
	default:
		if len(types) == 0 {
			fmt.Fprintln(w, "No types found.")
			return nil
		}
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "TYPE\tCOUNT")
		for _, t := range types {
			fmt.Fprintf(tw, "%s\t%d\n", t.Type, t.Count)
		}
		return tw.Flush()
	}
}

func outputRelationships(cmd *cobra.Command, rels []model.RelationshipCount) error {
	w := cmd.OutOrStdout()
	switch getFormat(cmd) {
	case "json":
		return writeJSON(w, rels)
	default:
		if len(rels) == 0 {
			fmt.Fprintln(w, "No relationships found.")
			return nil
		}
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "RELATIONSHIP\tCOUNT")
		for _, r := range rels {
			fmt.Fprintf(tw, "%s\t%d\n", r.Relationship, r.Count)
		}
		return tw.Flush()
	}
}

func outputStats(cmd *cobra.Command, stats *model.Stats) error {
	w := cmd.OutOrStdout()
	switch getFormat(cmd) {
	case "json":
		return writeJSON(w, stats)
	default:
		fmt.Fprintf(w, "Entities:  %d\n", stats.EntityCount)
		fmt.Fprintf(w, "Edges:     %d\n", stats.EdgeCount)
		fmt.Fprintf(w, "Vectors:   %d\n", stats.VectorCount)
		fmt.Fprintf(w, "Orphans:   %d\n", stats.OrphanCount)
		fmt.Fprintf(w, "Types:     %d\n", stats.TypeCount)
		if stats.DBSize > 0 {
			fmt.Fprintf(w, "DB Size:   %s\n", formatBytes(stats.DBSize))
		}
		return nil
	}
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatBytes(b int64) string {
	switch {
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
