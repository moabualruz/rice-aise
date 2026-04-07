//go:build integration

package integration_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/mkh/rice-aise/internal/domain"
	"github.com/mkh/rice-aise/internal/platform"
	"github.com/mkh/rice-aise/internal/service"
	"github.com/mkh/rice-aise/internal/storage"
)

const vanillaProfile = "vanilla"

type lifecycleFixture struct {
	home     string
	store    *storage.FileProfileStore
	service  *service.ProfileService
	registry []domain.Tool
}

func TestLifecycleIntegration(t *testing.T) {
	t.Run("init lifecycle", testInitLifecycle)
	t.Run("use switching preserves content", testUseSwitching)
	t.Run("use only switches selected tool", testUseOnly)
	t.Run("create and delete profile", testCreateAndDelete)
	t.Run("aider dotfile symlinks", testAiderDotfileSymlinks)
	t.Run("opencode four directory switch", testOpenCodeAtomicSwitch)
	t.Run("copilot dual directory switch", testCopilotSwitch)
}

func testInitLifecycle(t *testing.T) {
	fx := newLifecycleFixture(t)

	current, err := fx.service.Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	expectedCurrent := time.Now().Format("2006-01-02") + "-current"
	if current != expectedCurrent {
		t.Fatalf("Init() current profile = %q, want %q", current, expectedCurrent)
	}

	assertProfileNames(t, fx.store, current, vanillaProfile)
	assertAllLiveLinksPointToProfile(t, fx, current)
	assertActiveProfiles(t, fx, map[string]string{
		"claude":   current,
		"aider":    current,
		"opencode": current,
		"copilot":  current,
	})
	assertCurrentProfileCapturedOriginals(t, fx, current)
	assertVanillaProfileSeededCredentialsOnly(t, fx)

	if _, err := fx.service.Init(); err == nil || !strings.Contains(err.Error(), "already initialized") {
		t.Fatalf("second Init() error = %v, want already initialized", err)
	}
}

func testUseSwitching(t *testing.T) {
	fx := newLifecycleFixture(t)
	current := mustInit(t, fx)

	if err := fx.service.Use(vanillaProfile, nil); err != nil {
		t.Fatalf("Use(vanilla) error = %v", err)
	}
	assertAllLiveLinksPointToProfile(t, fx, vanillaProfile)
	assertActiveProfiles(t, fx, map[string]string{
		"claude":   vanillaProfile,
		"aider":    vanillaProfile,
		"opencode": vanillaProfile,
		"copilot":  vanillaProfile,
	})
	assertLiveVanillaState(t, fx)

	if err := fx.service.Use(current, nil); err != nil {
		t.Fatalf("Use(%q) error = %v", current, err)
	}
	assertAllLiveLinksPointToProfile(t, fx, current)
	assertActiveProfiles(t, fx, map[string]string{
		"claude":   current,
		"aider":    current,
		"opencode": current,
		"copilot":  current,
	})
	assertLiveOriginalState(t, fx)
}

func testUseOnly(t *testing.T) {
	fx := newLifecycleFixture(t)
	current := mustInit(t, fx)

	if err := fx.service.Use(vanillaProfile, []string{"claude"}); err != nil {
		t.Fatalf("Use(vanilla, only claude) error = %v", err)
	}

	assertDirLinkToProfile(t, fx, "claude", 0, vanillaProfile)
	assertDirLinkToProfile(t, fx, "aider", 0, current)
	assertDotfileLinkToProfile(t, fx, "aider", ".aider.conf.yml", current)
	assertDotfileLinkToProfile(t, fx, "aider", ".aider.env", current)
	for i := range findTool(t, fx, "opencode").ConfigDirs {
		assertDirLinkToProfile(t, fx, "opencode", i, current)
	}
	for i := range findTool(t, fx, "copilot").ConfigDirs {
		assertDirLinkToProfile(t, fx, "copilot", i, current)
	}

	assertActiveProfiles(t, fx, map[string]string{
		"claude":   vanillaProfile,
		"aider":    current,
		"opencode": current,
		"copilot":  current,
	})
}

func testCreateAndDelete(t *testing.T) {
	fx := newLifecycleFixture(t)
	current := mustInit(t, fx)

	if err := fx.service.Create("custom"); err != nil {
		t.Fatalf("Create(custom) error = %v", err)
	}

	assertProfileNames(t, fx.store, current, "custom", vanillaProfile)
	assertProfileExists(t, fx.store.ProfileDir("custom"))
	assertProfileHasCredentialOnlyCopies(t, fx, "custom")

	if err := fx.store.Delete("custom"); err != nil {
		t.Fatalf("Delete(custom) error = %v", err)
	}

	assertProfileNames(t, fx.store, current, vanillaProfile)
	assertProfileMissing(t, fx.store.ProfileDir("custom"))
}

func testAiderDotfileSymlinks(t *testing.T) {
	fx := newLifecycleFixture(t)
	current := mustInit(t, fx)

	assertDotfileLinkToProfile(t, fx, "aider", ".aider.conf.yml", current)
	assertDotfileLinkToProfile(t, fx, "aider", ".aider.env", current)

	if err := fx.service.Use(vanillaProfile, []string{"aider"}); err != nil {
		t.Fatalf("Use(vanilla, aider) error = %v", err)
	}
	assertDotfileLinkToProfile(t, fx, "aider", ".aider.conf.yml", vanillaProfile)
	assertDotfileLinkToProfile(t, fx, "aider", ".aider.env", vanillaProfile)

	if err := fx.service.Use(current, []string{"aider"}); err != nil {
		t.Fatalf("Use(%q, aider) error = %v", current, err)
	}
	assertDotfileLinkToProfile(t, fx, "aider", ".aider.conf.yml", current)
	assertDotfileLinkToProfile(t, fx, "aider", ".aider.env", current)
}

func testOpenCodeAtomicSwitch(t *testing.T) {
	fx := newLifecycleFixture(t)
	mustInit(t, fx)

	for i := range findTool(t, fx, "opencode").ConfigDirs {
		assertDirLinkToProfile(t, fx, "opencode", i, currentProfileName(t, fx, "opencode"))
	}

	if err := fx.service.Use(vanillaProfile, []string{"opencode"}); err != nil {
		t.Fatalf("Use(vanilla, opencode) error = %v", err)
	}
	for i := range findTool(t, fx, "opencode").ConfigDirs {
		assertDirLinkToProfile(t, fx, "opencode", i, vanillaProfile)
	}
}

func testCopilotSwitch(t *testing.T) {
	fx := newLifecycleFixture(t)
	mustInit(t, fx)

	for i := range findTool(t, fx, "copilot").ConfigDirs {
		assertDirLinkToProfile(t, fx, "copilot", i, currentProfileName(t, fx, "copilot"))
	}

	if err := fx.service.Use(vanillaProfile, []string{"copilot"}); err != nil {
		t.Fatalf("Use(vanilla, copilot) error = %v", err)
	}
	for i := range findTool(t, fx, "copilot").ConfigDirs {
		assertDirLinkToProfile(t, fx, "copilot", i, vanillaProfile)
	}
}

func newLifecycleFixture(t *testing.T) *lifecycleFixture {
	t.Helper()

	home := t.TempDir()
	registry := testRegistry()
	seedLiveFilesystem(t, home)

	store := storage.NewFileProfileStore(filepath.Join(home, ".raise"))
	resolver := platform.NewPathResolver(home)

	return &lifecycleFixture{
		home:     home,
		store:    store,
		service:  service.NewProfileService(store, resolver, registry),
		registry: registry,
	}
}

func testRegistry() []domain.Tool {
	return []domain.Tool{
		{
			ID:   "claude",
			Name: "Claude Style",
			ConfigDirs: []domain.DirMapping{
				{SourcePath: "~/.claude", ProfileSubdir: "claude"},
			},
			CredentialFiles: []string{".credentials.json"},
		},
		{
			ID:   "aider",
			Name: "Aider Style",
			ConfigDirs: []domain.DirMapping{
				{SourcePath: "~/.aider", ProfileSubdir: "aider-cache"},
			},
			CredentialFiles: []string{"~/.aider.env"},
			DotfilesPattern: true,
			Dotfiles:        []string{".aider.conf.yml", ".aider.env"},
		},
		{
			ID:   "opencode",
			Name: "OpenCode Style",
			ConfigDirs: []domain.DirMapping{
				{SourcePath: "~/.config/opencode", ProfileSubdir: "opencode-config"},
				{SourcePath: "~/.local/share/opencode", ProfileSubdir: "opencode-data"},
				{SourcePath: "~/.cache/opencode", ProfileSubdir: "opencode-cache"},
				{SourcePath: "~/.local/state/opencode", ProfileSubdir: "opencode-state"},
			},
			CredentialFiles: []string{"~/.local/share/opencode/auth.json"},
		},
		{
			ID:   "copilot",
			Name: "Copilot Style",
			ConfigDirs: []domain.DirMapping{
				{SourcePath: "~/.config/github-copilot", ProfileSubdir: "copilot-auth"},
				{SourcePath: "~/.copilot", ProfileSubdir: "copilot-config"},
			},
			CredentialFiles: []string{"~/.config/github-copilot/apps.json"},
		},
	}
}

func seedLiveFilesystem(t *testing.T, home string) {
	t.Helper()

	files := map[string]string{
		".claude/.credentials.json":             `{"token":"claude-original"}`,
		".claude/settings.json":                 `{"theme":"dark"}`,
		".aider/cache/history.txt":              "history\n",
		".aider.conf.yml":                       "model: sonnet\n",
		".aider.env":                            "AIDER_API_KEY=original\n",
		".config/opencode/config/settings.json": `{"editor":"vim"}`,
		".local/share/opencode/auth.json":       `{"token":"opencode-original"}`,
		".cache/opencode/cache.db":              "cache\n",
		".local/state/opencode/session.json":    `{"session":"abc"}`,
		".config/github-copilot/apps.json":      `{"token":"copilot-original"}`,
		".config/github-copilot/hosts.json":     `{"github.com":true}`,
		".copilot/config.json":                  `{"telemetry":false}`,
	}

	for rel, contents := range files {
		writeFile(t, filepath.Join(home, rel), contents)
	}
}

func mustInit(t *testing.T, fx *lifecycleFixture) string {
	t.Helper()

	current, err := fx.service.Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	return current
}

func assertAllLiveLinksPointToProfile(t *testing.T, fx *lifecycleFixture, profile string) {
	t.Helper()

	for _, tool := range fx.registry {
		for i := range tool.ConfigDirs {
			assertDirLinkToProfile(t, fx, tool.ID, i, profile)
		}
		if !tool.DotfilesPattern {
			continue
		}
		for _, dotfile := range tool.Dotfiles {
			assertDotfileLinkToProfile(t, fx, tool.ID, dotfile, profile)
		}
	}
}

func assertCurrentProfileCapturedOriginals(t *testing.T, fx *lifecycleFixture, profile string) {
	t.Helper()

	assertFileContent(t, filepath.Join(fx.store.ProfileDir(profile), "claude", ".credentials.json"), `{"token":"claude-original"}`)
	assertFileContent(t, filepath.Join(fx.store.ProfileDir(profile), "claude", "settings.json"), `{"theme":"dark"}`)
	assertFileContent(t, filepath.Join(fx.store.ProfileDir(profile), "aider-cache", "cache", "history.txt"), "history\n")
	assertFileContent(t, filepath.Join(fx.store.ProfileDir(profile), "aider", ".aider.conf.yml"), "model: sonnet\n")
	assertFileContent(t, filepath.Join(fx.store.ProfileDir(profile), "aider", ".aider.env"), "AIDER_API_KEY=original\n")
	assertFileContent(t, filepath.Join(fx.store.ProfileDir(profile), "opencode-config", "config", "settings.json"), `{"editor":"vim"}`)
	assertFileContent(t, filepath.Join(fx.store.ProfileDir(profile), "opencode-data", "auth.json"), `{"token":"opencode-original"}`)
	assertFileContent(t, filepath.Join(fx.store.ProfileDir(profile), "opencode-cache", "cache.db"), "cache\n")
	assertFileContent(t, filepath.Join(fx.store.ProfileDir(profile), "opencode-state", "session.json"), `{"session":"abc"}`)
	assertFileContent(t, filepath.Join(fx.store.ProfileDir(profile), "copilot-auth", "apps.json"), `{"token":"copilot-original"}`)
	assertFileContent(t, filepath.Join(fx.store.ProfileDir(profile), "copilot-auth", "hosts.json"), `{"github.com":true}`)
	assertFileContent(t, filepath.Join(fx.store.ProfileDir(profile), "copilot-config", "config.json"), `{"telemetry":false}`)
}

func assertVanillaProfileSeededCredentialsOnly(t *testing.T, fx *lifecycleFixture) {
	t.Helper()
	assertProfileHasCredentialOnlyCopies(t, fx, vanillaProfile)
}

func assertProfileHasCredentialOnlyCopies(t *testing.T, fx *lifecycleFixture, profile string) {
	t.Helper()

	root := fx.store.ProfileDir(profile)

	assertFileContent(t, filepath.Join(root, "claude", ".credentials.json"), `{"token":"claude-original"}`)
	assertNoEntry(t, filepath.Join(root, "claude", "settings.json"))

	assertFileContent(t, filepath.Join(root, "aider", ".aider.env"), "AIDER_API_KEY=original\n")
	assertFileContent(t, filepath.Join(root, "aider", ".aider.conf.yml"), "")
	assertNoEntry(t, filepath.Join(root, "aider-cache", "cache", "history.txt"))

	assertFileContent(t, filepath.Join(root, "opencode-data", "auth.json"), `{"token":"opencode-original"}`)
	assertNoEntry(t, filepath.Join(root, "opencode-config", "config", "settings.json"))
	assertNoEntry(t, filepath.Join(root, "opencode-cache", "cache.db"))
	assertNoEntry(t, filepath.Join(root, "opencode-state", "session.json"))

	assertFileContent(t, filepath.Join(root, "copilot-auth", "apps.json"), `{"token":"copilot-original"}`)
	assertNoEntry(t, filepath.Join(root, "copilot-auth", "hosts.json"))
	assertNoEntry(t, filepath.Join(root, "copilot-config", "config.json"))
}

func assertLiveVanillaState(t *testing.T, fx *lifecycleFixture) {
	t.Helper()

	assertFileContent(t, filepath.Join(fx.home, ".claude", ".credentials.json"), `{"token":"claude-original"}`)
	assertNoEntry(t, filepath.Join(fx.home, ".claude", "settings.json"))

	assertFileContent(t, filepath.Join(fx.home, ".aider.env"), "AIDER_API_KEY=original\n")
	assertFileContent(t, filepath.Join(fx.home, ".aider.conf.yml"), "")
	assertNoEntry(t, filepath.Join(fx.home, ".aider", "cache", "history.txt"))

	assertFileContent(t, filepath.Join(fx.home, ".local/share/opencode", "auth.json"), `{"token":"opencode-original"}`)
	assertNoEntry(t, filepath.Join(fx.home, ".config/opencode", "config", "settings.json"))
	assertNoEntry(t, filepath.Join(fx.home, ".cache/opencode", "cache.db"))
	assertNoEntry(t, filepath.Join(fx.home, ".local/state/opencode", "session.json"))

	assertFileContent(t, filepath.Join(fx.home, ".config/github-copilot", "apps.json"), `{"token":"copilot-original"}`)
	assertNoEntry(t, filepath.Join(fx.home, ".config/github-copilot", "hosts.json"))
	assertNoEntry(t, filepath.Join(fx.home, ".copilot", "config.json"))
}

func assertLiveOriginalState(t *testing.T, fx *lifecycleFixture) {
	t.Helper()

	assertFileContent(t, filepath.Join(fx.home, ".claude", ".credentials.json"), `{"token":"claude-original"}`)
	assertFileContent(t, filepath.Join(fx.home, ".claude", "settings.json"), `{"theme":"dark"}`)
	assertFileContent(t, filepath.Join(fx.home, ".aider", "cache", "history.txt"), "history\n")
	assertFileContent(t, filepath.Join(fx.home, ".aider.conf.yml"), "model: sonnet\n")
	assertFileContent(t, filepath.Join(fx.home, ".aider.env"), "AIDER_API_KEY=original\n")
	assertFileContent(t, filepath.Join(fx.home, ".config/opencode", "config", "settings.json"), `{"editor":"vim"}`)
	assertFileContent(t, filepath.Join(fx.home, ".local/share/opencode", "auth.json"), `{"token":"opencode-original"}`)
	assertFileContent(t, filepath.Join(fx.home, ".cache/opencode", "cache.db"), "cache\n")
	assertFileContent(t, filepath.Join(fx.home, ".local/state/opencode", "session.json"), `{"session":"abc"}`)
	assertFileContent(t, filepath.Join(fx.home, ".config/github-copilot", "apps.json"), `{"token":"copilot-original"}`)
	assertFileContent(t, filepath.Join(fx.home, ".config/github-copilot", "hosts.json"), `{"github.com":true}`)
	assertFileContent(t, filepath.Join(fx.home, ".copilot", "config.json"), `{"telemetry":false}`)
}

func assertActiveProfiles(t *testing.T, fx *lifecycleFixture, want map[string]string) {
	t.Helper()

	for toolID, expected := range want {
		got, err := fx.service.Which(toolID)
		if err != nil {
			t.Fatalf("Which(%q) error = %v", toolID, err)
		}
		if got != expected {
			t.Fatalf("Which(%q) = %q, want %q", toolID, got, expected)
		}
	}
}

func assertProfileNames(t *testing.T, store *storage.FileProfileStore, want ...string) {
	t.Helper()

	profiles, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	got := make([]string, 0, len(profiles))
	for _, profile := range profiles {
		got = append(got, profile.Name)
	}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("List() profiles = %v, want %v", got, want)
	}
}

func assertDirLinkToProfile(t *testing.T, fx *lifecycleFixture, toolID string, dirIndex int, profile string) {
	t.Helper()

	tool := findTool(t, fx, toolID)
	livePath := filepath.Join(fx.home, strings.TrimPrefix(tool.ConfigDirs[dirIndex].SourcePath, "~/"))
	target := filepath.Join(fx.store.ProfileDir(profile), tool.ConfigDirs[dirIndex].ProfileSubdir)
	assertSymlinkTarget(t, livePath, target)
}

func assertDotfileLinkToProfile(t *testing.T, fx *lifecycleFixture, toolID, dotfile, profile string) {
	t.Helper()

	livePath := filepath.Join(fx.home, dotfile)
	target := filepath.Join(fx.store.ProfileDir(profile), toolID, filepath.Clean(dotfile))
	assertSymlinkTarget(t, livePath, target)
}

func assertSymlinkTarget(t *testing.T, path, wantTarget string) {
	t.Helper()

	info, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("Lstat(%q) error = %v", path, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%q is not a symlink", path)
	}
	gotTarget, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("Readlink(%q) error = %v", path, err)
	}
	if filepath.Clean(gotTarget) != filepath.Clean(wantTarget) {
		t.Fatalf("Readlink(%q) = %q, want %q", path, gotTarget, wantTarget)
	}
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if string(data) != want {
		t.Fatalf("ReadFile(%q) = %q, want %q", path, string(data), want)
	}
}

func assertNoEntry(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Lstat(path); err == nil {
		t.Fatalf("expected %q to be absent", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("Lstat(%q) error = %v", path, err)
	}
}

func assertProfileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("Stat(%q) error = %v", path, err)
	}
}

func assertProfileMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected profile %q to be removed", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("Stat(%q) error = %v", path, err)
	}
}

func currentProfileName(t *testing.T, fx *lifecycleFixture, toolID string) string {
	t.Helper()
	name, err := fx.service.Which(toolID)
	if err != nil {
		t.Fatalf("Which(%q) error = %v", toolID, err)
	}
	return name
}

func findTool(t *testing.T, fx *lifecycleFixture, toolID string) domain.Tool {
	t.Helper()

	tool := domain.FindTool(fx.registry, toolID)
	if tool == nil {
		t.Fatalf("tool %q not found in registry", toolID)
	}
	return *tool
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
