package cli

import (
	"bufio"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

func newDeleteCommand(dep dependencies) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, dep, args[0], force)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Delete without confirmation")
	return cmd
}

func runDelete(cmd *cobra.Command, dep dependencies, name string, force bool) error {
	if !force {
		ok, err := confirmDelete(cmd.InOrStdin(), cmd.OutOrStdout())
		if err != nil || !ok {
			return finishDeletePrompt(cmd, err, ok)
		}
	}
	if err := dep.store.Delete(name); err != nil {
		return err
	}
	if err := removeActiveProfile(dep.store, name); err != nil {
		return err
	}
	return printLine(cmd.OutOrStdout(), "Deleted profile %q\n", name)
}

func finishDeletePrompt(cmd *cobra.Command, err error, ok bool) error {
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return printLine(cmd.OutOrStdout(), "Delete cancelled.\n")
}

func confirmDelete(r io.Reader, w io.Writer) (bool, error) {
	if err := printLine(w, "Are you sure? [y/N]: "); err != nil {
		return false, err
	}
	line, err := bufio.NewReader(r).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}

func removeActiveProfile(store profileStore, name string) error {
	active, err := store.GetActive()
	if err != nil {
		return err
	}
	if !dropActiveProfile(active, name) {
		return nil
	}
	return store.SetActive(active)
}

func dropActiveProfile(active map[string]string, name string) bool {
	changed := false
	for toolID, profile := range active {
		if profile != name {
			continue
		}
		delete(active, toolID)
		changed = true
	}
	return changed
}
