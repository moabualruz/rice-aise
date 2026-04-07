package cli

import (
	"strings"
	"testing"
)

func TestCommandsExecute(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		args         []string
		input        string
		wantContains []string
		check        func(*testing.T, *mockService, *mockStore)
	}{
		{
			name:         "init",
			args:         []string{"init"},
			wantContains: []string{"Initialized raise", "/tmp/.raise", "current"},
		},
		{
			name:         "use with only",
			args:         []string{"use", "work", "--only", "claude,codex"},
			wantContains: []string{"Using profile \"work\" for tools: claude, codex"},
			check: func(t *testing.T, svc *mockService, _ *mockStore) {
				t.Helper()
				if svc.useName != "work" {
					t.Fatalf("useName = %q", svc.useName)
				}
				if strings.Join(svc.useOnly, ",") != "claude,codex" {
					t.Fatalf("useOnly = %v", svc.useOnly)
				}
			},
		},
		{
			name:         "list",
			args:         []string{"list"},
			wantContains: []string{"* alpha", "* beta"},
		},
		{
			name:         "status",
			args:         []string{"status"},
			wantContains: []string{"Tool", "claude", "alpha", "codex", "beta"},
		},
		{
			name:         "create",
			args:         []string{"create", "new-profile"},
			wantContains: []string{"Created profile \"new-profile\""},
		},
		{
			name:         "save",
			args:         []string{"save", "snapshot"},
			wantContains: []string{"Saved active configuration as \"snapshot\""},
		},
		{
			name:         "rename",
			args:         []string{"rename", "alpha", "renamed"},
			wantContains: []string{"Renamed profile \"alpha\" to \"renamed\""},
			check: func(t *testing.T, _ *mockService, store *mockStore) {
				t.Helper()
				if store.renameOld != "alpha" || store.renameNew != "renamed" {
					t.Fatalf("rename = %q -> %q", store.renameOld, store.renameNew)
				}
				if store.active["claude"] != "renamed" {
					t.Fatalf("active claude = %q", store.active["claude"])
				}
			},
		},
		{
			name:         "delete force",
			args:         []string{"delete", "beta", "--force"},
			wantContains: []string{"Deleted profile \"beta\""},
			check: func(t *testing.T, _ *mockService, store *mockStore) {
				t.Helper()
				if store.deleted != "beta" {
					t.Fatalf("deleted = %q", store.deleted)
				}
				if _, ok := store.active["codex"]; ok {
					t.Fatal("codex active mapping should be removed")
				}
			},
		},
		{
			name:         "delete confirm yes",
			args:         []string{"delete", "beta"},
			input:        "y\n",
			wantContains: []string{"Are you sure? [y/N]: ", "Deleted profile \"beta\""},
		},
		{
			name:         "delete confirm no",
			args:         []string{"delete", "beta"},
			input:        "n\n",
			wantContains: []string{"Are you sure? [y/N]: ", "Delete cancelled."},
			check: func(t *testing.T, _ *mockService, store *mockStore) {
				t.Helper()
				if store.deleted != "" {
					t.Fatalf("delete should not run, got %q", store.deleted)
				}
			},
		},
		{
			name:         "tools",
			args:         []string{"tools"},
			wantContains: []string{"ID", "Name", "Config Dirs", "Credentials"},
		},
		{
			name:         "which",
			args:         []string{"which", "claude"},
			wantContains: []string{"claude -> alpha"},
		},
		{
			name:         "version",
			args:         []string{"version"},
			wantContains: []string{"raise v0.1.0"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dep, svc, store := newTestDeps()
			output, err := executeCommand(dep, tc.input, tc.args...)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			for _, want := range tc.wantContains {
				if !strings.Contains(output, want) {
					t.Fatalf("output %q missing %q", output, want)
				}
			}
			if tc.check != nil {
				tc.check(t, svc, store)
			}
		})
	}
}

func TestToolsOutputContainsRegistryIDs(t *testing.T) {
	t.Parallel()

	dep, _, _ := newTestDeps()
	output, err := executeCommand(dep, "", "tools")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	for _, tool := range dep.registry {
		if !strings.Contains(output, tool.ID) {
			t.Fatalf("tools output missing %q", tool.ID)
		}
	}
}

func TestVersionOutput(t *testing.T) {
	t.Parallel()

	dep, _, _ := newTestDeps()
	output, err := executeCommand(dep, "", "version")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(output, "v0.1.0") {
		t.Fatalf("output = %q", output)
	}
}

func TestCommandArgumentValidation(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "delete requires name", args: []string{"delete"}, wantErr: "accepts 1 arg(s), received 0"},
		{name: "rename requires old and new", args: []string{"rename", "alpha"}, wantErr: "accepts 2 arg(s), received 1"},
		{name: "save requires name", args: []string{"save"}, wantErr: "accepts 1 arg(s), received 0"},
		{name: "use requires name", args: []string{"use"}, wantErr: "accepts 1 arg(s), received 0"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dep, _, _ := newTestDeps()
			_, err := executeCommand(dep, "", tc.args...)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %v, want substring %q", err, tc.wantErr)
			}
		})
	}
}
