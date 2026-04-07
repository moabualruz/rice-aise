package platform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mkh/rice-aise/internal/domain"
)

func TestNewPathResolverUsesInjectedHomeDir(t *testing.T) {
	resolver := NewPathResolver("/tmp/custom-home")

	if got := resolver.HomeDir(); got != "/tmp/custom-home" {
		t.Fatalf("HomeDir() = %q, want %q", got, "/tmp/custom-home")
	}
}

func TestNewPathResolverUsesSystemHomeDirWhenEmpty(t *testing.T) {
	want, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() error = %v", err)
	}

	resolver := NewPathResolver("")
	if got := resolver.HomeDir(); got != want {
		t.Fatalf("HomeDir() = %q, want %q", got, want)
	}
}

func TestExpandTilde(t *testing.T) {
	resolver := NewPathResolver("/tmp/home")

	cases := []struct {
		name string
		path string
		want string
	}{
		{name: "exact tilde", path: "~", want: "/tmp/home"},
		{name: "tilde prefix", path: "~/config", want: filepath.Join("/tmp/home", "config")},
		{name: "non tilde", path: "/var/tmp/config", want: "/var/tmp/config"},
		{name: "empty", path: "", want: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolver.ExpandTilde(tc.path); got != tc.want {
				t.Fatalf("ExpandTilde(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}

func TestResolvePathCleansExpandedPath(t *testing.T) {
	resolver := NewPathResolver("/tmp/home")
	want := filepath.Clean(filepath.Join("/tmp/home", "bar"))

	if got := resolver.ResolvePath("~/foo/../bar"); got != want {
		t.Fatalf("ResolvePath() = %q, want %q", got, want)
	}
}

func TestExists(t *testing.T) {
	resolver := NewPathResolver("/tmp/home")
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "exists.txt")
	if err := os.WriteFile(existingFile, []byte("ok"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	if !resolver.Exists(existingFile) {
		t.Fatalf("Exists(%q) = false, want true", existingFile)
	}

	missingFile := filepath.Join(tempDir, "missing.txt")
	if resolver.Exists(missingFile) {
		t.Fatalf("Exists(%q) = true, want false", missingFile)
	}
}

func TestIsSymlink(t *testing.T) {
	resolver := NewPathResolver("/tmp/home")
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "target.txt")
	if err := os.WriteFile(target, []byte("ok"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	link := filepath.Join(tempDir, "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("os.Symlink() error = %v", err)
	}

	if !resolver.IsSymlink(link) {
		t.Fatalf("IsSymlink(%q) = false, want true", link)
	}
	if resolver.IsSymlink(target) {
		t.Fatalf("IsSymlink(%q) = true, want false", target)
	}

	missing := filepath.Join(tempDir, "missing-link.txt")
	if resolver.IsSymlink(missing) {
		t.Fatalf("IsSymlink(%q) = true, want false", missing)
	}
}

func TestResolveToolDirUsesEnvOverride(t *testing.T) {
	t.Setenv("TOOL_HOME", "/tmp/override/../tool")
	resolver := NewPathResolver("/tmp/home")
	tool := domain.Tool{
		EnvOverride: "TOOL_HOME",
		ConfigDirs: []domain.DirMapping{
			{SourcePath: "~/.tool"},
		},
	}

	want := filepath.Clean("/tmp/override/../tool")
	if got := resolver.ResolveToolDir(tool, 0); got != want {
		t.Fatalf("ResolveToolDir() = %q, want %q", got, want)
	}
}

func TestResolveToolDirFallsBackToSourcePath(t *testing.T) {
	resolver := NewPathResolver("/tmp/home")
	tool := domain.Tool{
		EnvOverride: "UNSET_TOOL_HOME",
		ConfigDirs: []domain.DirMapping{
			{SourcePath: "~/.tool"},
		},
	}

	want := filepath.Join("/tmp/home", ".tool")
	if got := resolver.ResolveToolDir(tool, 0); got != want {
		t.Fatalf("ResolveToolDir() = %q, want %q", got, want)
	}
}
