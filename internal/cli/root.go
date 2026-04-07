package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/mkh/rice-aise/internal/domain"
	"github.com/mkh/rice-aise/internal/platform"
	profilesvc "github.com/mkh/rice-aise/internal/service"
	"github.com/mkh/rice-aise/internal/storage"
	"github.com/spf13/cobra"
)

const Version = "0.1.0"

var (
	depsFactory           = newDeps
	exitFunc              = os.Exit
	stdin       io.Reader = os.Stdin
	stdout      io.Writer = os.Stdout
	stderr      io.Writer = os.Stderr
)

type profileService interface {
	DetectInstalled() []domain.Tool
	Init() (string, error)
	Save(name string) error
	Use(profileName string, onlyTools []string) error
	Create(name string) error
	Which(toolID string) (string, error)
}

type profileStore interface {
	List() ([]domain.Profile, error)
	Delete(name string) error
	Rename(oldName, newName string) error
	SetActive(mapping map[string]string) error
	GetActive() (map[string]string, error)
	RaiseDir() string
}

type dependencies struct {
	service  profileService
	store    profileStore
	registry []domain.Tool
}

func Execute() {
	dep, err := depsFactory()
	if err != nil {
		exitWithError(stderr, err)
	}
	cmd := newRootCommand(dep)
	cmd.SetIn(stdin)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	if err := cmd.Execute(); err != nil {
		exitWithError(stderr, err)
	}
}

func newDeps() (dependencies, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return dependencies{}, err
	}
	resolver := platform.NewPathResolver(homeDir)
	raiseDir := resolver.ResolvePath("~/.raise")
	store := storage.NewFileProfileStore(raiseDir)
	registry := domain.NewToolRegistry()
	service := profilesvc.NewProfileService(store, resolver, registry)
	return dependencies{service: service, store: store, registry: registry}, nil
}

func newRootCommand(dep dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "raise",
		Short:         "AI agent config profile switcher",
		Long:          "AI agent config profile switcher",
		Version:       Version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.AddCommand(newInitCommand(dep))
	cmd.AddCommand(newUseCommand(dep))
	cmd.AddCommand(newListCommand(dep))
	cmd.AddCommand(newStatusCommand(dep))
	cmd.AddCommand(newCreateCommand(dep))
	cmd.AddCommand(newSaveCommand(dep))
	cmd.AddCommand(newRenameCommand(dep))
	cmd.AddCommand(newDeleteCommand(dep))
	cmd.AddCommand(newToolsCommand(dep))
	cmd.AddCommand(newWhichCommand(dep))
	cmd.AddCommand(newVersionCommand())
	return cmd
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the raise version",
		Args:  cobra.NoArgs,
		RunE:  runVersion,
	}
}

func runVersion(cmd *cobra.Command, _ []string) error {
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s v%s\n", cmd.Root().Name(), Version)
	return err
}

func printLine(w io.Writer, format string, args ...any) error {
	_, err := fmt.Fprintf(w, format, args...)
	return err
}

func exitWithError(w io.Writer, err error) {
	_, _ = fmt.Fprintln(w, err)
	exitFunc(1)
}
