package domain

// NewToolRegistry returns the complete registry of supported AI CLI tools.
func NewToolRegistry() []Tool {
	return []Tool{
		claude(),
		codex(),
		gemini(),
		opencode(),
		pi(),
		copilot(),
		omc(),
		omx(),
		omg(),
		aider(),
		cline(),
		continueCLI(),
		amp(),
		goose(),
		amazonq(),
		kiro(),
		kilo(),
	}
}

// FindTool returns a tool by ID from the registry, or nil if not found.
func FindTool(registry []Tool, id string) *Tool {
	for i := range registry {
		if registry[i].ID == id {
			return &registry[i]
		}
	}
	return nil
}

// ToolIDs returns all tool IDs from the registry.
func ToolIDs(registry []Tool) []string {
	ids := make([]string, len(registry))
	for i, t := range registry {
		ids[i] = t.ID
	}
	return ids
}

func claude() Tool {
	return Tool{
		ID:   "claude",
		Name: "Claude Code",
		ConfigDirs: []DirMapping{
			{SourcePath: "~/.claude", ProfileSubdir: "claude"},
		},
		CredentialFiles: []string{".credentials.json"},
		EnvOverride:     "CLAUDE_CONFIG_DIR",
	}
}

func codex() Tool {
	return Tool{
		ID:   "codex",
		Name: "Codex CLI",
		ConfigDirs: []DirMapping{
			{SourcePath: "~/.codex", ProfileSubdir: "codex"},
		},
		CredentialFiles: []string{"auth.json"},
		EnvOverride:     "CODEX_HOME",
	}
}

func gemini() Tool {
	return Tool{
		ID:   "gemini",
		Name: "Gemini CLI",
		ConfigDirs: []DirMapping{
			{SourcePath: "~/.gemini", ProfileSubdir: "gemini"},
		},
		CredentialFiles: []string{
			"oauth_creds.json",
			"google_accounts.json",
			"mcp-oauth-tokens.json",
		},
		EnvOverride: "GEMINI_CLI_HOME",
	}
}

func opencode() Tool {
	return Tool{
		ID:   "opencode",
		Name: "OpenCode",
		ConfigDirs: []DirMapping{
			{SourcePath: "~/.config/opencode", ProfileSubdir: "opencode-config"},
			{SourcePath: "~/.local/share/opencode", ProfileSubdir: "opencode-data"},
			{SourcePath: "~/.cache/opencode", ProfileSubdir: "opencode-cache"},
			{SourcePath: "~/.local/state/opencode", ProfileSubdir: "opencode-state"},
		},
		CredentialFiles: []string{"~/.local/share/opencode/auth.json"},
		EnvOverride:     "OPENCODE_CONFIG_DIR",
	}
}

func pi() Tool {
	return Tool{
		ID:   "pi",
		Name: "Pi Coding Agent",
		ConfigDirs: []DirMapping{
			{SourcePath: "~/.pi", ProfileSubdir: "pi"},
		},
		CredentialFiles: []string{"agent/auth.json"},
		EnvOverride:     "PI_CONFIG_DIR",
	}
}

func copilot() Tool {
	return Tool{
		ID:   "copilot",
		Name: "GitHub Copilot CLI",
		ConfigDirs: []DirMapping{
			{SourcePath: "~/.config/github-copilot", ProfileSubdir: "copilot-auth"},
			{SourcePath: "~/.copilot", ProfileSubdir: "copilot-config"},
		},
		CredentialFiles: []string{"~/.config/github-copilot/apps.json"},
		EnvOverride:     "COPILOT_HOME",
	}
}

func omc() Tool {
	return Tool{
		ID:   "omc",
		Name: "oh-my-claudecode",
		ConfigDirs: []DirMapping{
			{SourcePath: "~/.omc", ProfileSubdir: "omc"},
		},
		CredentialFiles: nil,
	}
}

func omx() Tool {
	return Tool{
		ID:   "omx",
		Name: "oh-my-codex",
		ConfigDirs: []DirMapping{
			{SourcePath: "~/.omx", ProfileSubdir: "omx"},
		},
		CredentialFiles: nil,
	}
}

func omg() Tool {
	return Tool{
		ID:   "omg",
		Name: "oh-my-gemini",
		ConfigDirs: []DirMapping{
			{SourcePath: "~/.omg", ProfileSubdir: "omg"},
		},
		CredentialFiles: nil,
	}
}

func aider() Tool {
	return Tool{
		ID:   "aider",
		Name: "Aider",
		ConfigDirs: []DirMapping{
			{SourcePath: "~/.aider", ProfileSubdir: "aider-cache"},
		},
		CredentialFiles: []string{"~/.aider.env"},
		DotfilesPattern: true,
		Dotfiles: []string{
			".aider.conf.yml",
			".aider.model.settings.yml",
			".aider.model.metadata.json",
			".aider.env",
		},
	}
}

func cline() Tool {
	return Tool{
		ID:   "cline",
		Name: "Cline CLI",
		ConfigDirs: []DirMapping{
			{SourcePath: "~/.cline", ProfileSubdir: "cline"},
		},
		CredentialFiles: []string{"data/secrets.json"},
		EnvOverride:     "CLINE_DIR",
	}
}

func continueCLI() Tool {
	return Tool{
		ID:   "continue",
		Name: "Continue CLI",
		ConfigDirs: []DirMapping{
			{SourcePath: "~/.continue", ProfileSubdir: "continue"},
		},
		CredentialFiles: []string{"config.yaml"},
	}
}

func amp() Tool {
	return Tool{
		ID:   "amp",
		Name: "Amp",
		ConfigDirs: []DirMapping{
			{SourcePath: "~/.config/amp", ProfileSubdir: "amp"},
		},
		CredentialFiles: nil,
	}
}

func goose() Tool {
	return Tool{
		ID:   "goose",
		Name: "Goose",
		ConfigDirs: []DirMapping{
			{SourcePath: "~/.config/block/goose", ProfileSubdir: "goose-config"},
			{SourcePath: "~/.local/share/goose", ProfileSubdir: "goose-data"},
			{SourcePath: "~/.local/state/goose", ProfileSubdir: "goose-state"},
		},
		CredentialFiles: nil,
	}
}

func amazonq() Tool {
	return Tool{
		ID:   "amazonq",
		Name: "Amazon Q Developer CLI",
		ConfigDirs: []DirMapping{
			{SourcePath: "~/.aws/amazonq", ProfileSubdir: "amazonq"},
		},
		CredentialFiles: nil,
	}
}

func kiro() Tool {
	return Tool{
		ID:   "kiro",
		Name: "Kiro CLI",
		ConfigDirs: []DirMapping{
			{SourcePath: "~/.kiro", ProfileSubdir: "kiro"},
		},
		CredentialFiles: nil,
	}
}

func kilo() Tool {
	return Tool{
		ID:   "kilo",
		Name: "Kilo Code CLI",
		ConfigDirs: []DirMapping{
			{SourcePath: "~/.config/kilo", ProfileSubdir: "kilo"},
		},
		CredentialFiles: nil,
	}
}
