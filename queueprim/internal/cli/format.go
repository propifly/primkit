package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/propifly/primkit/queueprim/internal/model"
	"github.com/spf13/cobra"
)

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

func outputJob(cmd *cobra.Command, job *model.Job) error {
	w := cmd.OutOrStdout()
	switch getFormat(cmd) {
	case "json":
		return writeJSON(w, job)
	case "quiet":
		fmt.Fprintln(w, job.ID)
		return nil
	default:
		return writeJobDetail(w, job)
	}
}

func outputJobs(cmd *cobra.Command, jobs []*model.Job) error {
	w := cmd.OutOrStdout()
	switch getFormat(cmd) {
	case "json":
		return writeJSON(w, jobs)
	case "quiet":
		for _, j := range jobs {
			fmt.Fprintln(w, j.ID)
		}
		return nil
	default:
		return writeJobTable(w, jobs)
	}
}

func outputQueues(cmd *cobra.Command, queues []model.QueueInfo) error {
	w := cmd.OutOrStdout()
	switch getFormat(cmd) {
	case "json":
		return writeJSON(w, queues)
	default:
		if len(queues) == 0 {
			fmt.Fprintln(w, "No queues found.")
			return nil
		}
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "QUEUE\tPENDING\tCLAIMED\tDONE\tFAILED\tDEAD")
		for _, q := range queues {
			fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\t%d\n",
				q.Queue, q.Pending, q.Claimed, q.Done, q.Failed, q.Dead)
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
		total := stats.TotalPending + stats.TotalClaimed + stats.TotalDone + stats.TotalFailed + stats.TotalDead
		fmt.Fprintf(w, "Pending: %d\n", stats.TotalPending)
		fmt.Fprintf(w, "Claimed: %d\n", stats.TotalClaimed)
		fmt.Fprintf(w, "Done:    %d\n", stats.TotalDone)
		fmt.Fprintf(w, "Failed:  %d\n", stats.TotalFailed)
		fmt.Fprintf(w, "Dead:    %d\n", stats.TotalDead)
		fmt.Fprintf(w, "Total:   %d\n", total)
		return nil
	}
}

func writeJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func writeJobDetail(w io.Writer, j *model.Job) error {
	fmt.Fprintf(w, "ID:           %s\n", j.ID)
	fmt.Fprintf(w, "Queue:        %s\n", j.Queue)
	if j.Type != "" {
		fmt.Fprintf(w, "Type:         %s\n", j.Type)
	}
	fmt.Fprintf(w, "Priority:     %s\n", j.Priority)
	fmt.Fprintf(w, "Status:       %s\n", j.Status)
	fmt.Fprintf(w, "AttemptCount: %d / MaxRetries: %d\n", j.AttemptCount, j.MaxRetries)
	fmt.Fprintf(w, "Payload:      %s\n", truncate(string(j.Payload), 80))
	if j.ClaimedBy != nil {
		fmt.Fprintf(w, "ClaimedBy:    %s\n", *j.ClaimedBy)
	}
	if j.ClaimedAt != nil {
		fmt.Fprintf(w, "ClaimedAt:    %s\n", j.ClaimedAt.Format(time.RFC3339))
	}
	fmt.Fprintf(w, "VisibleAfter: %s\n", j.VisibleAfter.Format(time.RFC3339))
	if j.CompletedAt != nil {
		fmt.Fprintf(w, "CompletedAt:  %s\n", j.CompletedAt.Format(time.RFC3339))
	}
	if j.FailureReason != nil {
		fmt.Fprintf(w, "FailureReason:%s\n", *j.FailureReason)
	}
	if len(j.Output) > 0 {
		fmt.Fprintf(w, "Output:       %s\n", truncate(string(j.Output), 80))
	}
	fmt.Fprintf(w, "CreatedAt:    %s\n", j.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "UpdatedAt:    %s\n", j.UpdatedAt.Format(time.RFC3339))
	return nil
}

func writeJobTable(w io.Writer, jobs []*model.Job) error {
	if len(jobs) == 0 {
		fmt.Fprintln(w, "No jobs found.")
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tQUEUE\tTYPE\tPRIORITY\tSTATUS\tATTEMPTS")
	for _, j := range jobs {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d\n",
			j.ID,
			truncate(j.Queue, 30),
			j.Type,
			j.Priority,
			j.Status,
			j.AttemptCount,
		)
	}
	return tw.Flush()
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
