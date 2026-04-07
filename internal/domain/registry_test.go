package domain

import (
	"testing"
)

func TestNewToolRegistryReturns17Tools(t *testing.T) {
	registry := NewToolRegistry()
	if len(registry) != 17 {
		t.Errorf("expected 17 tools, got %d", len(registry))
	}
}

func TestAllToolsHaveRequiredFields(t *testing.T) {
	registry := NewToolRegistry()
	for _, tool := range registry {
		if tool.ID == "" {
			t.Error("tool has empty ID")
		}
		if tool.Name == "" {
			t.Errorf("tool %s has empty Name", tool.ID)
		}
		if len(tool.ConfigDirs) == 0 && !tool.DotfilesPattern {
			t.Errorf("tool %s has no ConfigDirs and is not a dotfiles pattern", tool.ID)
		}
		if tool.DotfilesPattern && len(tool.Dotfiles) == 0 {
			t.Errorf("tool %s is dotfiles pattern but has no Dotfiles listed", tool.ID)
		}
	}
}

func TestAllToolIDsAreUnique(t *testing.T) {
	registry := NewToolRegistry()
	seen := make(map[string]bool)
	for _, tool := range registry {
		if seen[tool.ID] {
			t.Errorf("duplicate tool ID: %s", tool.ID)
		}
		seen[tool.ID] = true
	}
}

func TestClaudeTool(t *testing.T) {
	tool := FindTool(NewToolRegistry(), "claude")
	if tool == nil {
		t.Fatal("claude tool not found")
	}
	assertDirCount(t, tool, 1)
	assertDirSource(t, tool, 0, "~/.claude")
	assertCredFile(t, tool, ".credentials.json")
	assertEqual(t, "EnvOverride", tool.EnvOverride, "CLAUDE_CONFIG_DIR")
}

func TestCodexTool(t *testing.T) {
	tool := FindTool(NewToolRegistry(), "codex")
	if tool == nil {
		t.Fatal("codex tool not found")
	}
	assertDirCount(t, tool, 1)
	assertDirSource(t, tool, 0, "~/.codex")
	assertCredFile(t, tool, "auth.json")
	assertEqual(t, "EnvOverride", tool.EnvOverride, "CODEX_HOME")
}

func TestGeminiTool(t *testing.T) {
	tool := FindTool(NewToolRegistry(), "gemini")
	if tool == nil {
		t.Fatal("gemini tool not found")
	}
	assertDirCount(t, tool, 1)
	assertDirSource(t, tool, 0, "~/.gemini")
	assertCredFiles(t, tool, []string{"oauth_creds.json", "google_accounts.json", "mcp-oauth-tokens.json"})
	assertEqual(t, "EnvOverride", tool.EnvOverride, "GEMINI_CLI_HOME")
}

func TestOpenCodeToolHas4Dirs(t *testing.T) {
	tool := FindTool(NewToolRegistry(), "opencode")
	if tool == nil {
		t.Fatal("opencode tool not found")
	}
	assertDirCount(t, tool, 4)

	expectedSources := []string{
		"~/.config/opencode",
		"~/.local/share/opencode",
		"~/.cache/opencode",
		"~/.local/state/opencode",
	}
	for i, expected := range expectedSources {
		assertDirSource(t, tool, i, expected)
	}
	assertCredFile(t, tool, "~/.local/share/opencode/auth.json")
}

func TestPiTool(t *testing.T) {
	tool := FindTool(NewToolRegistry(), "pi")
	if tool == nil {
		t.Fatal("pi tool not found")
	}
	assertDirCount(t, tool, 1)
	assertDirSource(t, tool, 0, "~/.pi")
	assertCredFile(t, tool, "agent/auth.json")
	assertEqual(t, "EnvOverride", tool.EnvOverride, "PI_CONFIG_DIR")
}

func TestCopilotToolHas2Dirs(t *testing.T) {
	tool := FindTool(NewToolRegistry(), "copilot")
	if tool == nil {
		t.Fatal("copilot tool not found")
	}
	assertDirCount(t, tool, 2)
	assertDirSource(t, tool, 0, "~/.config/github-copilot")
	assertDirSource(t, tool, 1, "~/.copilot")
	assertCredFile(t, tool, "~/.config/github-copilot/apps.json")
}

func TestOMCTool(t *testing.T) {
	tool := FindTool(NewToolRegistry(), "omc")
	if tool == nil {
		t.Fatal("omc tool not found")
	}
	assertDirCount(t, tool, 1)
	assertDirSource(t, tool, 0, "~/.omc")
	if len(tool.CredentialFiles) != 0 {
		t.Errorf("omc should have no credential files, got %v", tool.CredentialFiles)
	}
}

func TestOMXTool(t *testing.T) {
	tool := FindTool(NewToolRegistry(), "omx")
	if tool == nil {
		t.Fatal("omx tool not found")
	}
	assertDirCount(t, tool, 1)
	assertDirSource(t, tool, 0, "~/.omx")
	if len(tool.CredentialFiles) != 0 {
		t.Errorf("omx should have no credential files, got %v", tool.CredentialFiles)
	}
}

func TestOMGTool(t *testing.T) {
	tool := FindTool(NewToolRegistry(), "omg")
	if tool == nil {
		t.Fatal("omg tool not found")
	}
	assertDirCount(t, tool, 1)
	assertDirSource(t, tool, 0, "~/.omg")
	if len(tool.CredentialFiles) != 0 {
		t.Errorf("omg should have no credential files, got %v", tool.CredentialFiles)
	}
}

func TestAiderToolIsDotfilesPattern(t *testing.T) {
	tool := FindTool(NewToolRegistry(), "aider")
	if tool == nil {
		t.Fatal("aider tool not found")
	}
	if !tool.DotfilesPattern {
		t.Error("aider should have DotfilesPattern=true")
	}
	// 1 dir (~/.aider/) + 4 dotfiles
	if len(tool.Dotfiles) < 4 {
		t.Errorf("aider should have at least 4 dotfiles, got %d", len(tool.Dotfiles))
	}
	assertDirCount(t, tool, 1)
	assertDirSource(t, tool, 0, "~/.aider")
	assertCredFile(t, tool, "~/.aider.env")
}

func TestClineTool(t *testing.T) {
	tool := FindTool(NewToolRegistry(), "cline")
	if tool == nil {
		t.Fatal("cline tool not found")
	}
	assertDirCount(t, tool, 1)
	assertDirSource(t, tool, 0, "~/.cline")
	assertCredFile(t, tool, "data/secrets.json")
	assertEqual(t, "EnvOverride", tool.EnvOverride, "CLINE_DIR")
}

func TestContinueTool(t *testing.T) {
	tool := FindTool(NewToolRegistry(), "continue")
	if tool == nil {
		t.Fatal("continue tool not found")
	}
	assertDirCount(t, tool, 1)
	assertDirSource(t, tool, 0, "~/.continue")
	assertCredFile(t, tool, "config.yaml")
}

func TestAmpTool(t *testing.T) {
	tool := FindTool(NewToolRegistry(), "amp")
	if tool == nil {
		t.Fatal("amp tool not found")
	}
	assertDirCount(t, tool, 1)
	assertDirSource(t, tool, 0, "~/.config/amp")
}

func TestGooseTool(t *testing.T) {
	tool := FindTool(NewToolRegistry(), "goose")
	if tool == nil {
		t.Fatal("goose tool not found")
	}
	assertDirCount(t, tool, 3)
	assertDirSource(t, tool, 0, "~/.config/block/goose")
	assertDirSource(t, tool, 1, "~/.local/share/goose")
	assertDirSource(t, tool, 2, "~/.local/state/goose")
}

func TestAmazonQToolDoesNotManageAWSRoot(t *testing.T) {
	tool := FindTool(NewToolRegistry(), "amazonq")
	if tool == nil {
		t.Fatal("amazonq tool not found")
	}
	assertDirCount(t, tool, 1)
	assertDirSource(t, tool, 0, "~/.aws/amazonq")
	// Must NOT include ~/.aws/ itself
	for _, dir := range tool.ConfigDirs {
		if dir.SourcePath == "~/.aws" {
			t.Error("amazonq must NOT manage ~/.aws/ directly — only ~/.aws/amazonq/")
		}
	}
}

func TestKiroTool(t *testing.T) {
	tool := FindTool(NewToolRegistry(), "kiro")
	if tool == nil {
		t.Fatal("kiro tool not found")
	}
	assertDirCount(t, tool, 1)
	assertDirSource(t, tool, 0, "~/.kiro")
}

func TestKiloTool(t *testing.T) {
	tool := FindTool(NewToolRegistry(), "kilo")
	if tool == nil {
		t.Fatal("kilo tool not found")
	}
	assertDirCount(t, tool, 1)
	assertDirSource(t, tool, 0, "~/.config/kilo")
}

func TestFindToolReturnsNilForUnknown(t *testing.T) {
	tool := FindTool(NewToolRegistry(), "nonexistent")
	if tool != nil {
		t.Error("expected nil for unknown tool")
	}
}

func TestToolIDsReturnsAll(t *testing.T) {
	ids := ToolIDs(NewToolRegistry())
	if len(ids) != 17 {
		t.Errorf("expected 17 tool IDs, got %d", len(ids))
	}
}

// --- test helpers ---

func assertDirCount(t *testing.T, tool *Tool, expected int) {
	t.Helper()
	if len(tool.ConfigDirs) != expected {
		t.Errorf("tool %s: expected %d config dirs, got %d", tool.ID, expected, len(tool.ConfigDirs))
	}
}

func assertDirSource(t *testing.T, tool *Tool, idx int, expected string) {
	t.Helper()
	if idx >= len(tool.ConfigDirs) {
		t.Errorf("tool %s: config dir index %d out of range (len=%d)", tool.ID, idx, len(tool.ConfigDirs))
		return
	}
	if tool.ConfigDirs[idx].SourcePath != expected {
		t.Errorf("tool %s: config dir[%d] source = %q, want %q", tool.ID, idx, tool.ConfigDirs[idx].SourcePath, expected)
	}
}

func assertCredFile(t *testing.T, tool *Tool, expected string) {
	t.Helper()
	for _, f := range tool.CredentialFiles {
		if f == expected {
			return
		}
	}
	t.Errorf("tool %s: expected credential file %q not found in %v", tool.ID, expected, tool.CredentialFiles)
}

func assertCredFiles(t *testing.T, tool *Tool, expected []string) {
	t.Helper()
	for _, e := range expected {
		assertCredFile(t, tool, e)
	}
}

func assertEqual(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %q, want %q", field, got, want)
	}
}
