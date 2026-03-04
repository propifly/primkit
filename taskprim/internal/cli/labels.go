package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLabelsCmd() *cobra.Command {
	var list string

	cmd := &cobra.Command{
		Use:   "labels",
		Short: "List labels with open task counts",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			labels, err := s.ListLabels(cmd.Context(), list)
			if err != nil {
				return fmt.Errorf("listing labels: %w", err)
			}

			return outputLabels(cmd, labels)
		},
	}

	cmd.Flags().StringVarP(&list, "list", "l", "", "only labels from tasks in this list")

	// Add the "clear" subcommand.
	cmd.AddCommand(newLabelClearCmd())

	return cmd
}

func newLabelClearCmd() *cobra.Command {
	var list string

	cmd := &cobra.Command{
		Use:   "clear <label>",
		Short: "Remove a label from all tasks",
		Long: `Removes a label from every task that carries it. If --list is specified,
only tasks in that list are affected.

This is commonly used in morning-brief flows:
  taskprim labels clear today --list andres`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())
			label := args[0]

			cleared, err := s.ClearLabel(cmd.Context(), label, list)
			if err != nil {
				return fmt.Errorf("clearing label: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Cleared %q from %d task(s)\n", label, cleared)
			return nil
		},
	}

	cmd.Flags().StringVarP(&list, "list", "l", "", "only clear from tasks in this list")

	return cmd
}
