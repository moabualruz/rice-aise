package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newStatusCommand(dep dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show active profiles by tool",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runStatus(cmd, dep)
		},
	}
}

func runStatus(cmd *cobra.Command, dep dependencies) error {
	active, err := dep.store.GetActive()
	if err != nil {
		return err
	}
	table := newTabWriter(cmd)
	writeStatusHeader(table)
	for _, tool := range dep.registry {
		if err := writeStatusRow(table, tool.ID, active[tool.ID]); err != nil {
			return err
		}
	}
	return table.Flush()
}

func writeStatusHeader(table *tabwriter.Writer) {
	fmt.Fprintln(table, "Tool\tActive Profile")
}

func writeStatusRow(table *tabwriter.Writer, toolID, profile string) error {
	if profile == "" {
		profile = "-"
	}
	_, err := fmt.Fprintf(table, "%s\t%s\n", toolID, profile)
	return err
}
