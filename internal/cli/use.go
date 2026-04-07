package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

func newUseCommand(dep dependencies) *cobra.Command {
	var onlyTools []string
	cmd := &cobra.Command{
		Use:   "use <name>",
		Short: "Switch to a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUse(cmd, dep, args[0], onlyTools)
		},
	}
	cmd.Flags().StringSliceVar(&onlyTools, "only", nil, "Only switch selected tools")
	return cmd
}

func runUse(
	cmd *cobra.Command,
	dep dependencies,
	name string,
	onlyTools []string,
) error {
	if err := dep.service.Use(name, onlyTools); err != nil {
		return err
	}
	if len(onlyTools) == 0 {
		return printLine(cmd.OutOrStdout(), "Using profile %q for all tools\n", name)
	}
	return printLine(
		cmd.OutOrStdout(),
		"Using profile %q for tools: %s\n",
		name,
		strings.Join(onlyTools, ", "),
	)
}
