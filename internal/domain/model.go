package domain

import "time"

// DirMapping maps a source path on disk to a subdirectory name within a profile.
type DirMapping struct {
	SourcePath    string
	ProfileSubdir string
}

// Tool represents an AI CLI agent with its configuration layout.
type Tool struct {
	ID              string
	Name            string
	ConfigDirs      []DirMapping
	CredentialFiles []string // relative to respective config dir, or absolute
	EnvOverride     string
	DotfilesPattern bool     // true if tool uses home-dir dotfiles instead of a directory
	Dotfiles        []string // individual dotfile paths (relative to home) when DotfilesPattern=true
}

// Profile represents a named configuration snapshot.
type Profile struct {
	Name      string
	CreatedAt time.Time
	Tools     map[string]bool // tool ID -> whether included in this profile
}
