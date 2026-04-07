package cli

import "github.com/spf13/cobra"

func newCreateCommand(dep dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new profile from current credentials",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, dep, args[0])
		},
	}
}

func runCreate(cmd *cobra.Command, dep dependencies, name string) error {
	if err := dep.service.Create(name); err != nil {
		return err
	}
	return printLine(cmd.OutOrStdout(), "Created profile %q\n", name)
}
