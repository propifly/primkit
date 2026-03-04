package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newNamespacesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "namespaces",
		Short: "List all namespaces with record counts",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := storeFromContext(cmd.Context())

			nss, err := s.ListNamespaces(cmd.Context())
			if err != nil {
				return fmt.Errorf("listing namespaces: %w", err)
			}

			return outputNamespaces(cmd, nss)
		},
	}
}
