//go:build e2e

package e2e_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var (
	binaryPath string
	repoRoot   string
)

func TestMain(m *testing.M) {
	repoRoot = findRepoRoot()
	tmpBinDir, err := os.MkdirTemp("", "raise-e2e-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	binaryPath = filepath.Join(tmpBinDir, "raise")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/raise")
	buildCmd.Dir = repoRoot
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n%s", err, output)
		_ = os.RemoveAll(tmpBinDir)
		os.Exit(1)
	}

	code := m.Run()
	_ = os.RemoveAll(tmpBinDir)
	os.Exit(code)
}

func TestVersion(t *testing.T) {
	stdout, stderr, exitCode := runRaise(t, t.TempDir(), "version")
	if exitCode != 0 {
		t.Fatalf("version exit = %d, stderr = %q", exitCode, stderr)
	}
	requireContains(t, stdout, "raise v0.1.0")
}

func TestTools(t *testing.T) {
	stdout, stderr, exitCode := runRaise(t, t.TempDir(), "tools")
	if exitCode != 0 {
		t.Fatalf("tools exit = %d, stderr = %q", exitCode, stderr)
	}
	requireContains(t, stdout, "ID", "claude", "Claude Code", "codex", "Codex CLI", "gemini")
}

func TestInitAndList(t *testing.T) {
	homeDir := setupHomeWithClaude(t)
	current := mustInitRaise(t, homeDir)

	stdout, stderr, exitCode := runRaise(t, homeDir, "list")
	if exitCode != 0 {
		t.Fatalf("list exit = %d, stderr = %q", exitCode, stderr)
	}
	requireContains(t, stdout, current, "vanilla")
}

func TestStatus(t *testing.T) {
	homeDir := setupHomeWithClaude(t)
	current := mustInitRaise(t, homeDir)

	stdout, stderr, exitCode := runRaise(t, homeDir, "status")
	if exitCode != 0 {
		t.Fatalf("status exit = %d, stderr = %q", exitCode, stderr)
	}
	requireContains(t, stdout, "Tool", "Active Profile", "claude", current)
}

func TestUseAndSwitch(t *testing.T) {
	homeDir := setupHomeWithClaude(t)
	original := mustInitRaise(t, homeDir)

	stdout, stderr, exitCode := runRaise(t, homeDir, "use", "vanilla")
	if exitCode != 0 {
		t.Fatalf("use vanilla exit = %d, stderr = %q", exitCode, stderr)
	}
	requireContains(t, stdout, `Using profile "vanilla" for all tools`)

	stdout, stderr, exitCode = runRaise(t, homeDir, "status")
	if exitCode != 0 {
		t.Fatalf("status after vanilla exit = %d, stderr = %q", exitCode, stderr)
	}
	requireContains(t, stdout, "claude", "vanilla")

	stdout, stderr, exitCode = runRaise(t, homeDir, "use", original)
	if exitCode != 0 {
		t.Fatalf("use original exit = %d, stderr = %q", exitCode, stderr)
	}
	requireContains(t, stdout, fmt.Sprintf(`Using profile %q for all tools`, original))

	stdout, stderr, exitCode = runRaise(t, homeDir, "status")
	if exitCode != 0 {
		t.Fatalf("status after restore exit = %d, stderr = %q", exitCode, stderr)
	}
	requireContains(t, stdout, "claude", original)
}

func TestCreate(t *testing.T) {
	homeDir := setupHomeWithClaude(t)
	mustInitRaise(t, homeDir)

	stdout, stderr, exitCode := runRaise(t, homeDir, "create", "custom")
	if exitCode != 0 {
		t.Fatalf("create exit = %d, stderr = %q", exitCode, stderr)
	}
	requireContains(t, stdout, `Created profile "custom"`)

	stdout, stderr, exitCode = runRaise(t, homeDir, "list")
	if exitCode != 0 {
		t.Fatalf("list exit = %d, stderr = %q", exitCode, stderr)
	}
	requireContains(t, stdout, "custom")
}

func TestWhich(t *testing.T) {
	homeDir := setupHomeWithClaude(t)
	original := mustInitRaise(t, homeDir)

	stdout, stderr, exitCode := runRaise(t, homeDir, "which", "claude")
	if exitCode != 0 {
		t.Fatalf("which exit = %d, stderr = %q", exitCode, stderr)
	}
	requireContains(t, stdout, "claude -> "+original)
}

func TestDeleteWithForce(t *testing.T) {
	homeDir := setupHomeWithClaude(t)
	mustInitRaise(t, homeDir)

	stdout, stderr, exitCode := runRaise(t, homeDir, "create", "todelete")
	if exitCode != 0 {
		t.Fatalf("create todelete exit = %d, stderr = %q", exitCode, stderr)
	}
	requireContains(t, stdout, `Created profile "todelete"`)

	stdout, stderr, exitCode = runRaise(t, homeDir, "use", "todelete")
	if exitCode != 0 {
		t.Fatalf("use todelete exit = %d, stderr = %q", exitCode, stderr)
	}
	requireContains(t, stdout, `Using profile "todelete" for all tools`)

	stdout, stderr, exitCode = runRaise(t, homeDir, "use", "vanilla")
	if exitCode != 0 {
		t.Fatalf("use vanilla exit = %d, stderr = %q", exitCode, stderr)
	}

	stdout, stderr, exitCode = runRaise(t, homeDir, "delete", "todelete", "--force")
	if exitCode != 0 {
		t.Fatalf("delete --force exit = %d, stderr = %q", exitCode, stderr)
	}
	requireContains(t, stdout, `Deleted profile "todelete"`)

	stdout, stderr, exitCode = runRaise(t, homeDir, "list")
	if exitCode != 0 {
		t.Fatalf("list after delete exit = %d, stderr = %q", exitCode, stderr)
	}
	if strings.Contains(stdout, "todelete") {
		t.Fatalf("list still contains deleted profile: %q", stdout)
	}
}

func TestDeleteActiveProfileFails(t *testing.T) {
	homeDir := setupHomeWithClaude(t)
	original := mustInitRaise(t, homeDir)

	stdout, stderr, exitCode := runRaise(t, homeDir, "delete", original)
	if exitCode == 0 && !strings.Contains(stdout, "Delete cancelled.") {
		t.Fatalf("delete active unexpectedly succeeded: stdout = %q, stderr = %q", stdout, stderr)
	}
	if exitCode != 0 && !strings.Contains(stdout+stderr, original) {
		t.Fatalf("delete active failed without useful output: stdout = %q, stderr = %q", stdout, stderr)
	}
	requireContains(t, stdout, "Are you sure? [y/N]: ")

	stdout, stderr, exitCode = runRaise(t, homeDir, "which", "claude")
	if exitCode != 0 {
		t.Fatalf("which after failed delete exit = %d, stderr = %q", exitCode, stderr)
	}
	requireContains(t, stdout, "claude -> "+original)
}

func runRaise(t *testing.T, homeDir string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = repoRoot
	cmd.Env = envWithHome(homeDir)

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err == nil {
		return outBuf.String(), errBuf.String(), 0
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return outBuf.String(), errBuf.String(), exitErr.ExitCode()
	}

	t.Fatalf("run raise %q: %v", strings.Join(args, " "), err)
	return "", "", 0
}

func findRepoRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Fprintln(os.Stderr, "runtime.Caller failed")
		os.Exit(1)
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func envWithHome(homeDir string) []string {
	env := make([]string, 0, len(os.Environ())+1)
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "HOME=") {
			env = append(env, entry)
		}
	}
	return append(env, "HOME="+homeDir)
}

func setupHomeWithClaude(t *testing.T) string {
	t.Helper()

	homeDir := t.TempDir()
	credPath := filepath.Join(homeDir, ".claude", ".credentials.json")
	if err := os.MkdirAll(filepath.Dir(credPath), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(credPath), err)
	}
	if err := os.WriteFile(credPath, []byte(`{"token":"test"}`), 0o644); err != nil {
		t.Fatalf("write %s: %v", credPath, err)
	}
	return homeDir
}

func mustInitRaise(t *testing.T, homeDir string) string {
	t.Helper()

	stdout, stderr, exitCode := runRaise(t, homeDir, "init")
	if exitCode != 0 {
		t.Fatalf("init exit = %d, stderr = %q", exitCode, stderr)
	}
	requireContains(t, stdout, "Initialized raise in ", `current profile "`)

	stdout, stderr, exitCode = runRaise(t, homeDir, "which", "claude")
	if exitCode != 0 {
		t.Fatalf("which after init exit = %d, stderr = %q", exitCode, stderr)
	}
	prefix := "claude -> "
	if !strings.HasPrefix(stdout, prefix) {
		t.Fatalf("unexpected which output: %q", stdout)
	}
	return strings.TrimSpace(strings.TrimPrefix(stdout, prefix))
}

func requireContains(t *testing.T, output string, substrings ...string) {
	t.Helper()
	for _, substring := range substrings {
		if !strings.Contains(output, substring) {
			t.Fatalf("output %q does not contain %q", output, substring)
		}
	}
}
