package cli

import "github.com/spf13/cobra"

func newRenameCommand(dep dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "rename <old> <new>",
		Short: "Rename a profile",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRename(cmd, dep, args[0], args[1])
		},
	}
}

func runRename(cmd *cobra.Command, dep dependencies, oldName, newName string) error {
	if err := dep.store.Rename(oldName, newName); err != nil {
		return err
	}
	if err := renameActiveProfiles(dep.store, oldName, newName); err != nil {
		return err
	}
	return printLine(cmd.OutOrStdout(), "Renamed profile %q to %q\n", oldName, newName)
}

func renameActiveProfiles(store profileStore, oldName, newName string) error {
	active, err := store.GetActive()
	if err != nil {
		return err
	}
	if !replaceActiveProfile(active, oldName, newName) {
		return nil
	}
	return store.SetActive(active)
}

func replaceActiveProfile(active map[string]string, oldName, newName string) bool {
	changed := false
	for toolID, profile := range active {
		if profile != oldName {
			continue
		}
		active[toolID] = newName
		changed = true
	}
	return changed
}
