package cli

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/mkh/rice-aise/internal/domain"
	"github.com/spf13/cobra"
)

func newToolsCommand(dep dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "tools",
		Short: "List supported tools",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTools(cmd, dep)
		},
	}
}

func runTools(cmd *cobra.Command, dep dependencies) error {
	table := newTabWriter(cmd)
	fmt.Fprintln(table, "ID\tName\tConfig Dirs\tCredentials")
	for _, tool := range dep.registry {
		if err := writeToolRow(table, tool); err != nil {
			return err
		}
	}
	return table.Flush()
}

func writeToolRow(table *tabwriter.Writer, tool domain.Tool) error {
	_, err := fmt.Fprintf(
		table,
		"%s\t%s\t%s\t%s\n",
		tool.ID,
		tool.Name,
		joinConfigDirs(tool),
		joinOrDash(tool.CredentialFiles),
	)
	return err
}

func joinConfigDirs(tool domain.Tool) string {
	paths := make([]string, 0, len(tool.ConfigDirs))
	for _, dir := range tool.ConfigDirs {
		paths = append(paths, dir.SourcePath)
	}
	return joinOrDash(paths)
}

func joinOrDash(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ", ")
}

func newTabWriter(cmd *cobra.Command) *tabwriter.Writer {
	return tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
}
