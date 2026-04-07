package service

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mkh/rice-aise/internal/domain"
)

type mockStore struct {
	raiseDir     string
	profiles     map[string]*domain.Profile
	active       map[string]string
	listErr      error
	createErr    error
	failCreateAt int
	createCalls  int
	getActiveErr error
	setActiveErr error
}

func newMockStore(raiseDir string) *mockStore {
	return &mockStore{
		raiseDir: raiseDir,
		profiles: map[string]*domain.Profile{},
		active:   map[string]string{},
	}
}

func (s *mockStore) Create(name string, tools map[string]bool) (*domain.Profile, error) {
	s.createCalls++
	if s.failCreateAt > 0 && s.createCalls == s.failCreateAt {
		return nil, s.createErr
	}
	if s.createErr != nil && s.failCreateAt == 0 {
		return nil, s.createErr
	}
	profile := &domain.Profile{Name: name, CreatedAt: time.Now(), Tools: cloneBoolMap(tools)}
	s.profiles[name] = profile
	return profile, nil
}

func (s *mockStore) Get(name string) (*domain.Profile, error) {
	profile, ok := s.profiles[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	clone := *profile
	clone.Tools = cloneBoolMap(profile.Tools)
	return &clone, nil
}

func (s *mockStore) List() ([]domain.Profile, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	list := make([]domain.Profile, 0, len(s.profiles))
	for _, profile := range s.profiles {
		list = append(list, domain.Profile{Name: profile.Name, Tools: cloneBoolMap(profile.Tools)})
	}
	return list, nil
}

func (s *mockStore) Delete(name string) error             { delete(s.profiles, name); return nil }
func (s *mockStore) Rename(oldName, newName string) error { return nil }
func (s *mockStore) ProfileDir(name string) string {
	return filepath.Join(s.raiseDir, "profiles", name)
}
func (s *mockStore) RaiseDir() string { return s.raiseDir }
func (s *mockStore) SetActive(mapping map[string]string) error {
	if s.setActiveErr != nil {
		return s.setActiveErr
	}
	s.active = cloneStringMap(mapping)
	return nil
}

func (s *mockStore) GetActive() (map[string]string, error) {
	if s.getActiveErr != nil {
		return nil, s.getActiveErr
	}
	return cloneStringMap(s.active), nil
}

type mockResolver struct {
	homeDir  string
	exists   map[string]bool
	symlinks map[string]bool
}

func (r *mockResolver) HomeDir() string { return r.homeDir }

func (r *mockResolver) ExpandTilde(path string) string {
	if path == "~" {
		return r.homeDir
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(r.homeDir, path[2:])
	}
	return path
}

func (r *mockResolver) ResolvePath(path string) string { return filepath.Clean(r.ExpandTilde(path)) }
func (r *mockResolver) ResolveToolDir(tool domain.Tool, i int) string {
	return r.ResolvePath(tool.ConfigDirs[i].SourcePath)
}

func (r *mockResolver) Exists(path string) bool {
	if value, ok := r.exists[path]; ok {
		return value
	}
	_, err := os.Stat(path)
	return err == nil
}

func (r *mockResolver) IsSymlink(path string) bool {
	if value, ok := r.symlinks[path]; ok {
		return value
	}
	info, err := os.Lstat(path)
	return err == nil && info.Mode()&os.ModeSymlink != 0
}

func newHarness(t *testing.T) (string, string, *ProfileService, *mockStore) {
	t.Helper()
	home := t.TempDir()
	raiseDir := t.TempDir()
	store := newMockStore(raiseDir)
	resolver := &mockResolver{homeDir: home, exists: map[string]bool{}, symlinks: map[string]bool{}}
	service := NewProfileService(store, resolver, []domain.Tool{standardTool(), aiderTool(), multiTool()})
	return home, raiseDir, service, store
}

func standardTool() domain.Tool {
	return domain.Tool{
		ID:              "std",
		Name:            "Standard",
		ConfigDirs:      []domain.DirMapping{{SourcePath: "~/.std", ProfileSubdir: "std"}},
		CredentialFiles: []string{"auth.json"},
	}
}

func aiderTool() domain.Tool {
	return domain.Tool{
		ID:              "aider",
		Name:            "Aider",
		ConfigDirs:      []domain.DirMapping{{SourcePath: "~/.aider", ProfileSubdir: "aider-cache"}},
		CredentialFiles: []string{"~/.aider.env"},
		DotfilesPattern: true,
		Dotfiles:        []string{".aider.conf.yml", ".aider.env"},
	}
}

func multiTool() domain.Tool {
	return domain.Tool{
		ID:   "multi",
		Name: "Multi",
		ConfigDirs: []domain.DirMapping{
			{SourcePath: "~/.config/multi", ProfileSubdir: "multi-config"},
			{SourcePath: "~/.local/share/multi", ProfileSubdir: "multi-data"},
		},
		CredentialFiles: []string{"~/.local/share/multi/auth.json"},
	}
}

func setNow(t *testing.T, ts time.Time) func() {
	t.Helper()
	previous := nowFunc
	nowFunc = func() time.Time { return ts }
	return func() { nowFunc = previous }
}

func writeLiveStandard(t *testing.T, home string) {
	t.Helper()
	writeFile(t, filepath.Join(home, ".std", "config.txt"), "std-config")
	writeFile(t, filepath.Join(home, ".std", "auth.json"), "std-auth")
}

func writeLiveAider(t *testing.T, home string) {
	t.Helper()
	writeFile(t, filepath.Join(home, ".aider", "cache.txt"), "aider-cache")
	writeFile(t, filepath.Join(home, ".aider.conf.yml"), "aider-conf")
	writeFile(t, filepath.Join(home, ".aider.env"), "aider-env")
}

func writeLiveMulti(t *testing.T, home string) {
	t.Helper()
	writeFile(t, filepath.Join(home, ".config", "multi", "settings.yml"), "multi-config")
	writeFile(t, filepath.Join(home, ".local", "share", "multi", "auth.json"), "multi-auth")
	writeFile(t, filepath.Join(home, ".local", "share", "multi", "data.txt"), "multi-data")
}

func createProfileData(t *testing.T, profileDir, prefix string) {
	t.Helper()
	writeFile(t, filepath.Join(profileDir, "std", "config.txt"), prefix+"-std")
	writeFile(t, filepath.Join(profileDir, "aider-cache", "cache.txt"), prefix+"-cache")
	writeFile(t, filepath.Join(profileDir, "aider", ".aider.conf.yml"), prefix+"-conf")
	writeFile(t, filepath.Join(profileDir, "aider", ".aider.env"), prefix+"-env")
	writeFile(t, filepath.Join(profileDir, "multi-config", "settings.yml"), prefix+"-multi")
	writeFile(t, filepath.Join(profileDir, "multi-data", "auth.json"), prefix+"-auth")
}

func linkProfileToHome(t *testing.T, home, profileDir string) {
	t.Helper()
	mustSymlink(t, filepath.Join(profileDir, "std"), filepath.Join(home, ".std"))
	mustSymlink(t, filepath.Join(profileDir, "aider-cache"), filepath.Join(home, ".aider"))
	mustSymlink(t, filepath.Join(profileDir, "aider", ".aider.conf.yml"), filepath.Join(home, ".aider.conf.yml"))
	mustSymlink(t, filepath.Join(profileDir, "aider", ".aider.env"), filepath.Join(home, ".aider.env"))
	mustSymlink(t, filepath.Join(profileDir, "multi-config"), filepath.Join(home, ".config", "multi"))
	mustSymlink(t, filepath.Join(profileDir, "multi-data"), filepath.Join(home, ".local", "share", "multi"))
}

func mustSymlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(link)
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
}

func idsOf(tools []domain.Tool) []string {
	ids := make([]string, 0, len(tools))
	for _, tool := range tools {
		ids = append(ids, tool.ID)
	}
	return ids
}

func cloneBoolMap(src map[string]bool) map[string]bool {
	dst := map[string]bool{}
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func cloneStringMap(src map[string]string) map[string]string {
	dst := map[string]string{}
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if got := string(data); got != want {
		t.Fatalf("ReadFile(%q) = %q, want %q", path, got, want)
	}
}

func assertDir(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		t.Fatalf("Stat(%q) = %v, want directory", path, err)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Lstat(%q) error = %v, want not-exist", path, err)
	}
}

func assertLink(t *testing.T, link, wantTarget string) {
	t.Helper()
	got, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("Readlink(%q) error = %v", link, err)
	}
	if got != wantTarget {
		t.Fatalf("Readlink(%q) = %q, want %q", link, got, wantTarget)
	}
}
