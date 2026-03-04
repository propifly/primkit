package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/propifly/primkit/stateprim/internal/model"
	"github.com/spf13/cobra"
)

// getFormat extracts the --format flag value from the command chain.
func getFormat(cmd *cobra.Command) string {
	f, _ := cmd.Flags().GetString("format")
	if f == "" {
		f, _ = cmd.InheritedFlags().GetString("format")
	}
	if f == "" {
		return "table"
	}
	return f
}

// outputRecord formats a single record for display.
func outputRecord(cmd *cobra.Command, r *model.Record) error {
	w := cmd.OutOrStdout()
	switch getFormat(cmd) {
	case "json":
		return writeJSON(w, r)
	case "quiet":
		fmt.Fprintf(w, "%s/%s\n", r.Namespace, r.Key)
		return nil
	default:
		return writeRecordDetail(w, r)
	}
}

// outputRecords formats a list of records for display.
func outputRecords(cmd *cobra.Command, records []*model.Record) error {
	w := cmd.OutOrStdout()
	switch getFormat(cmd) {
	case "json":
		return writeJSON(w, records)
	case "quiet":
		for _, r := range records {
			fmt.Fprintf(w, "%s/%s\n", r.Namespace, r.Key)
		}
		return nil
	default:
		return writeRecordTable(w, records)
	}
}

// writeJSON outputs any value as indented JSON.
func writeJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// writeRecordDetail prints a single record in a detailed, human-readable format.
func writeRecordDetail(w io.Writer, r *model.Record) error {
	fmt.Fprintf(w, "Namespace:  %s\n", r.Namespace)
	fmt.Fprintf(w, "Key:        %s\n", r.Key)
	fmt.Fprintf(w, "Value:      %s\n", string(r.Value))
	fmt.Fprintf(w, "Immutable:  %v\n", r.Immutable)
	fmt.Fprintf(w, "Created:    %s\n", r.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "Updated:    %s\n", r.UpdatedAt.Format(time.RFC3339))
	return nil
}

// writeRecordTable prints records as an aligned table.
func writeRecordTable(w io.Writer, records []*model.Record) error {
	if len(records) == 0 {
		fmt.Fprintln(w, "No records found.")
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAMESPACE\tKEY\tIMMUTABLE\tUPDATED\tVALUE")
	for _, r := range records {
		val := truncate(string(r.Value), 50)
		fmt.Fprintf(tw, "%s\t%s\t%v\t%s\t%s\n",
			r.Namespace, r.Key, r.Immutable,
			r.UpdatedAt.Format(time.RFC3339), val)
	}
	return tw.Flush()
}

// outputNamespaces formats namespace info for display.
func outputNamespaces(cmd *cobra.Command, nss []model.NamespaceInfo) error {
	w := cmd.OutOrStdout()
	switch getFormat(cmd) {
	case "json":
		return writeJSON(w, nss)
	default:
		if len(nss) == 0 {
			fmt.Fprintln(w, "No namespaces found.")
			return nil
		}
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "NAMESPACE\tCOUNT")
		for _, ns := range nss {
			fmt.Fprintf(tw, "%s\t%d\n", ns.Namespace, ns.Count)
		}
		return tw.Flush()
	}
}

// outputStats formats stats for display.
func outputStats(cmd *cobra.Command, stats *model.Stats) error {
	w := cmd.OutOrStdout()
	switch getFormat(cmd) {
	case "json":
		return writeJSON(w, stats)
	default:
		fmt.Fprintf(w, "Records:    %d\n", stats.TotalRecords)
		fmt.Fprintf(w, "Namespaces: %d\n", stats.TotalNamespaces)
		return nil
	}
}

// truncate shortens a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
