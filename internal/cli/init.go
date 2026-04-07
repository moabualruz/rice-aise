package cli

import "github.com/spf13/cobra"

func newInitCommand(dep dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize raise in ~/.raise",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInit(cmd, dep)
		},
	}
}

func runInit(cmd *cobra.Command, dep dependencies) error {
	name, err := dep.service.Init()
	if err != nil {
		return err
	}
	return printLine(
		cmd.OutOrStdout(),
		"Initialized raise in %s with current profile %q\n",
		dep.store.RaiseDir(),
		name,
	)
}
