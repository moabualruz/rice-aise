package platform

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mkh/rice-aise/internal/domain"
)

// PathResolver abstracts platform-specific path handling for tool directories.
type PathResolver interface {
	HomeDir() string
	ExpandTilde(path string) string
	ResolvePath(path string) string
	Exists(path string) bool
	IsSymlink(path string) bool
	ResolveToolDir(tool domain.Tool, dirIndex int) string
}

// DefaultPathResolver provides filesystem-backed path resolution.
type DefaultPathResolver struct {
	homeDir string
}

// NewPathResolver creates a filesystem-backed resolver with an injectable home dir.
func NewPathResolver(homeDir string) PathResolver {
	if homeDir == "" {
		homeDir, _ = os.UserHomeDir()
	}
	return DefaultPathResolver{homeDir: homeDir}
}

// HomeDir returns the configured home directory.
func (r DefaultPathResolver) HomeDir() string {
	return r.homeDir
}

// ExpandTilde replaces a leading ~ with the configured home directory.
func (r DefaultPathResolver) ExpandTilde(path string) string {
	if path == "~" {
		return r.homeDir
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(r.homeDir, path[2:])
	}
	return path
}

// ResolvePath expands home-relative paths and normalizes separators.
func (r DefaultPathResolver) ResolvePath(path string) string {
	return filepath.Clean(r.ExpandTilde(path))
}

// Exists reports whether a filesystem entry exists at path.
func (r DefaultPathResolver) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsSymlink reports whether path points to a symbolic link.
func (r DefaultPathResolver) IsSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// ResolveToolDir resolves a tool config directory, preferring env overrides.
func (r DefaultPathResolver) ResolveToolDir(tool domain.Tool, dirIndex int) string {
	override := os.Getenv(tool.EnvOverride)
	if override != "" {
		return filepath.Clean(override)
	}
	return r.ResolvePath(tool.ConfigDirs[dirIndex].SourcePath)
}
