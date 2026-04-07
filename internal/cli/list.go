package cli

import (
	"sort"

	"github.com/mkh/rice-aise/internal/domain"
	"github.com/spf13/cobra"
)

func newListCommand(dep dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List profiles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runList(cmd, dep)
		},
	}
}

func runList(cmd *cobra.Command, dep dependencies) error {
	profiles, err := dep.store.List()
	if err != nil {
		return err
	}
	active, err := dep.store.GetActive()
	if err != nil {
		return err
	}
	sortProfiles(profiles)
	if len(profiles) == 0 {
		return printLine(cmd.OutOrStdout(), "No profiles found.\n")
	}
	for _, profile := range profiles {
		marker := activeMarker(active, profile.Name)
		if err := printLine(cmd.OutOrStdout(), "%s %s\n", marker, profile.Name); err != nil {
			return err
		}
	}
	return nil
}

func sortProfiles(profiles []domain.Profile) {
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})
}

func activeMarker(active map[string]string, profileName string) string {
	for _, current := range active {
		if current == profileName {
			return "*"
		}
	}
	return " "
}
