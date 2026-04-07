package cli

import "github.com/spf13/cobra"

func newWhichCommand(dep dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "which <tool-id>",
		Short: "Show which profile is active for a tool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhich(cmd, dep, args[0])
		},
	}
}

func runWhich(cmd *cobra.Command, dep dependencies, toolID string) error {
	name, err := dep.service.Which(toolID)
	if err != nil {
		return err
	}
	return printLine(cmd.OutOrStdout(), "%s -> %s\n", toolID, name)
}
