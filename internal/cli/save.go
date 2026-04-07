package cli

import "github.com/spf13/cobra"

func newSaveCommand(dep dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "save <name>",
		Short: "Save the active configuration as a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSave(cmd, dep, args[0])
		},
	}
}

func runSave(cmd *cobra.Command, dep dependencies, name string) error {
	if err := dep.service.Save(name); err != nil {
		return err
	}
	return printLine(cmd.OutOrStdout(), "Saved active configuration as %q\n", name)
}
