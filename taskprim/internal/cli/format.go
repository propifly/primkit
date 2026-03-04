package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/propifly/primkit/taskprim/internal/model"
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

// outputTask formats a single task for display.
func outputTask(cmd *cobra.Command, task *model.Task) error {
	w := cmd.OutOrStdout()
	switch getFormat(cmd) {
	case "json":
		return writeJSON(w, task)
	case "quiet":
		fmt.Fprintln(w, task.ID)
		return nil
	default:
		return writeTaskDetail(w, task)
	}
}

// outputTasks formats a list of tasks for display.
func outputTasks(cmd *cobra.Command, tasks []*model.Task) error {
	w := cmd.OutOrStdout()
	switch getFormat(cmd) {
	case "json":
		return writeJSON(w, tasks)
	case "quiet":
		for _, t := range tasks {
			fmt.Fprintln(w, t.ID)
		}
		return nil
	default:
		return writeTaskTable(w, tasks)
	}
}

// writeJSON outputs any value as indented JSON.
func writeJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// writeTaskDetail prints a single task in a detailed, human-readable format.
func writeTaskDetail(w io.Writer, t *model.Task) error {
	fmt.Fprintf(w, "ID:       %s\n", t.ID)
	fmt.Fprintf(w, "List:     %s\n", t.List)
	fmt.Fprintf(w, "What:     %s\n", t.What)
	fmt.Fprintf(w, "State:    %s\n", t.State)
	fmt.Fprintf(w, "Source:   %s\n", t.Source)
	if len(t.Labels) > 0 {
		fmt.Fprintf(w, "Labels:   %s\n", strings.Join(t.Labels, ", "))
	}
	if t.WaitingOn != nil {
		fmt.Fprintf(w, "Waiting:  %s\n", *t.WaitingOn)
	}
	if t.Context != nil {
		fmt.Fprintf(w, "Context:  %s\n", *t.Context)
	}
	if t.ParentID != nil {
		fmt.Fprintf(w, "Parent:   %s\n", *t.ParentID)
	}
	fmt.Fprintf(w, "Created:  %s\n", t.Created.Format(time.RFC3339))
	fmt.Fprintf(w, "Updated:  %s\n", t.Updated.Format(time.RFC3339))
	if t.ResolvedAt != nil {
		fmt.Fprintf(w, "Resolved: %s\n", t.ResolvedAt.Format(time.RFC3339))
	}
	if t.ResolvedReason != nil {
		fmt.Fprintf(w, "Reason:   %s\n", *t.ResolvedReason)
	}
	return nil
}

// writeTaskTable prints a list of tasks as an aligned table.
func writeTaskTable(w io.Writer, tasks []*model.Task) error {
	if len(tasks) == 0 {
		fmt.Fprintln(w, "No tasks found.")
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tLIST\tSTATE\tWHAT\tLABELS")
	for _, t := range tasks {
		labels := strings.Join(t.Labels, ", ")
		what := truncate(t.What, 50)
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", t.ID, t.List, t.State, what, labels)
	}
	return tw.Flush()
}

// outputLabels formats label counts for display.
func outputLabels(cmd *cobra.Command, labels []model.LabelCount) error {
	w := cmd.OutOrStdout()
	switch getFormat(cmd) {
	case "json":
		return writeJSON(w, labels)
	default:
		if len(labels) == 0 {
			fmt.Fprintln(w, "No labels found.")
			return nil
		}
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "LABEL\tCOUNT")
		for _, l := range labels {
			fmt.Fprintf(tw, "%s\t%d\n", l.Label, l.Count)
		}
		return tw.Flush()
	}
}

// outputLists formats list info for display.
func outputLists(cmd *cobra.Command, lists []model.ListInfo) error {
	w := cmd.OutOrStdout()
	switch getFormat(cmd) {
	case "json":
		return writeJSON(w, lists)
	default:
		if len(lists) == 0 {
			fmt.Fprintln(w, "No lists found.")
			return nil
		}
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "LIST\tOPEN\tDONE\tKILLED")
		for _, l := range lists {
			fmt.Fprintf(tw, "%s\t%d\t%d\t%d\n", l.Name, l.Open, l.Done, l.Killed)
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
		fmt.Fprintf(w, "Open:   %d\n", stats.TotalOpen)
		fmt.Fprintf(w, "Done:   %d\n", stats.TotalDone)
		fmt.Fprintf(w, "Killed: %d\n", stats.TotalKilled)
		fmt.Fprintf(w, "Total:  %d\n", stats.TotalOpen+stats.TotalDone+stats.TotalKilled)
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
