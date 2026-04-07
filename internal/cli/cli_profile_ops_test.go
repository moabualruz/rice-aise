package cli

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func TestRunInitError(t *testing.T) {
	t.Parallel()

	dep, svc, _ := newTestDeps()
	svc.initErr = errors.New("init failed")

	_, err := executeCommand(dep, "", "init")
	if !errors.Is(err, svc.initErr) {
		t.Fatalf("error = %v", err)
	}
}

func TestRunListPaths(t *testing.T) {
	t.Run("list returns no profiles message", func(t *testing.T) {
		t.Parallel()

		dep, _, store := newTestDeps()
		store.profiles = nil
		store.active = nil

		output, err := executeCommand(dep, "", "list")
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(output, "No profiles found.") {
			t.Fatalf("output = %q", output)
		}
	})

	t.Run("list propagates list error", func(t *testing.T) {
		t.Parallel()

		dep, _, store := newTestDeps()
		store.listErr = errors.New("list failed")

		_, err := executeCommand(dep, "", "list")
		if !errors.Is(err, store.listErr) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("list propagates active profile error", func(t *testing.T) {
		t.Parallel()

		dep, _, store := newTestDeps()
		store.activeErr = errors.New("active lookup failed")

		_, err := executeCommand(dep, "", "list")
		if !errors.Is(err, store.activeErr) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("list returns writer error", func(t *testing.T) {
		t.Parallel()

		dep, _, _ := newTestDeps()
		cmd := &cobra.Command{}
		cmd.SetOut(errWriter{err: errors.New("write failed")})

		err := runList(cmd, dep)
		if err == nil || !strings.Contains(err.Error(), "write failed") {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestActiveMarker(t *testing.T) {
	t.Parallel()

	if got := activeMarker(map[string]string{"claude": "alpha"}, "alpha"); got != "*" {
		t.Fatalf("activeMarker(match) = %q", got)
	}
	if got := activeMarker(map[string]string{"claude": "alpha"}, "beta"); got != " " {
		t.Fatalf("activeMarker(no match) = %q", got)
	}
}

func TestRunRenameErrors(t *testing.T) {
	t.Run("rename propagates store error", func(t *testing.T) {
		t.Parallel()

		dep, _, store := newTestDeps()
		store.renameErr = errors.New("rename failed")

		_, err := executeCommand(dep, "", "rename", "alpha", "renamed")
		if !errors.Is(err, store.renameErr) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("rename propagates active profile update error", func(t *testing.T) {
		t.Parallel()

		dep, _, store := newTestDeps()
		store.setErr = errors.New("set active failed")

		_, err := executeCommand(dep, "", "rename", "alpha", "renamed")
		if !errors.Is(err, store.setErr) {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestRenameActiveProfilesError(t *testing.T) {
	t.Parallel()

	store := &mockStore{
		active: map[string]string{"claude": "alpha"},
		setErr: errors.New("set active failed"),
	}

	err := renameActiveProfiles(store, "alpha", "beta")
	if !errors.Is(err, store.setErr) {
		t.Fatalf("error = %v", err)
	}
}

func TestRenameActiveProfilesGetActiveError(t *testing.T) {
	t.Parallel()

	store := &mockStore{activeErr: errors.New("active lookup failed")}
	if err := renameActiveProfiles(store, "alpha", "beta"); !errors.Is(err, store.activeErr) {
		t.Fatalf("error = %v", err)
	}
}

func TestRenameActiveProfilesNoChange(t *testing.T) {
	t.Parallel()

	store := &mockStore{active: map[string]string{"claude": "beta"}}
	if err := renameActiveProfiles(store, "alpha", "gamma"); err != nil {
		t.Fatalf("renameActiveProfiles() error = %v", err)
	}
	if store.active["claude"] != "beta" {
		t.Fatalf("active profile = %q", store.active["claude"])
	}
}

func TestRunSaveError(t *testing.T) {
	t.Parallel()

	dep, svc, _ := newTestDeps()
	svc.saveErr = errors.New("save failed")

	_, err := executeCommand(dep, "", "save", "snapshot")
	if !errors.Is(err, svc.saveErr) {
		t.Fatalf("error = %v", err)
	}
}

func TestRunStatusPaths(t *testing.T) {
	t.Run("status propagates active lookup error", func(t *testing.T) {
		t.Parallel()

		dep, _, store := newTestDeps()
		store.activeErr = errors.New("active lookup failed")

		_, err := executeCommand(dep, "", "status")
		if !errors.Is(err, store.activeErr) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("status returns writer error", func(t *testing.T) {
		t.Parallel()

		dep, _, _ := newTestDeps()
		cmd := &cobra.Command{}
		cmd.SetOut(&failAfterWriter{failOn: 2, err: errors.New("write failed")})

		err := runStatus(cmd, dep)
		if err == nil || !strings.Contains(err.Error(), "write failed") {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestWriteStatusRow(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		profile string
		want    string
	}{
		{name: "enabled profile", profile: "alpha", want: "claude  alpha"},
		{name: "disabled profile", profile: "", want: "claude  -"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			table := tabwriter.NewWriter(buf, 0, 0, 2, ' ', 0)
			if err := writeStatusRow(table, "claude", tc.profile); err != nil {
				t.Fatalf("writeStatusRow() error = %v", err)
			}
			if err := table.Flush(); err != nil {
				t.Fatalf("Flush() error = %v", err)
			}
			if !strings.Contains(buf.String(), tc.want) {
				t.Fatalf("output = %q, want substring %q", buf.String(), tc.want)
			}
		})
	}
}

func TestRunToolsPaths(t *testing.T) {
	t.Run("tools prints header with empty registry", func(t *testing.T) {
		t.Parallel()

		dep, _, _ := newTestDeps()
		dep.registry = nil

		output, err := executeCommand(dep, "", "tools")
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(output, "ID") {
			t.Fatalf("output = %q", output)
		}
	})

	t.Run("tools returns writer error", func(t *testing.T) {
		t.Parallel()

		dep, _, _ := newTestDeps()
		cmd := &cobra.Command{}
		cmd.SetOut(&failAfterWriter{failOn: 2, err: errors.New("write failed")})

		err := runTools(cmd, dep)
		if err == nil || !strings.Contains(err.Error(), "write failed") {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestRunUsePaths(t *testing.T) {
	t.Run("use without only flag targets all tools", func(t *testing.T) {
		t.Parallel()

		dep, svc, _ := newTestDeps()

		output, err := executeCommand(dep, "", "use", "work")
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(output, "Using profile \"work\" for all tools") {
			t.Fatalf("output = %q", output)
		}
		if len(svc.useOnly) != 0 {
			t.Fatalf("useOnly = %v", svc.useOnly)
		}
	})

	t.Run("use propagates service error", func(t *testing.T) {
		t.Parallel()

		dep, svc, _ := newTestDeps()
		svc.useErr = errors.New("use failed")

		_, err := executeCommand(dep, "", "use", "work")
		if !errors.Is(err, svc.useErr) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("use returns profile not found", func(t *testing.T) {
		t.Parallel()

		dep, svc, _ := newTestDeps()
		svc.useErr = errors.New("profile not found")

		_, err := executeCommand(dep, "", "use", "missing")
		if err == nil || !strings.Contains(err.Error(), "profile not found") {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestRunWhichErrors(t *testing.T) {
	t.Run("which propagates lookup error", func(t *testing.T) {
		t.Parallel()

		dep, svc, _ := newTestDeps()
		svc.whichErr = errors.New("which failed")

		_, err := executeCommand(dep, "", "which", "claude")
		if !errors.Is(err, svc.whichErr) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("which returns no active profile error", func(t *testing.T) {
		t.Parallel()

		dep, svc, _ := newTestDeps()
		svc.whichErr = errors.New("no active profile")

		_, err := executeCommand(dep, "", "which", "claude")
		if err == nil || !strings.Contains(err.Error(), "no active profile") {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestCommandErrorsAndHelpers(t *testing.T) {
	t.Run("command returns service error", func(t *testing.T) {
		dep, svc, _ := newTestDeps()
		svc.createErr = errors.New("boom")
		if _, err := executeCommand(dep, "", "create", "bad"); !errors.Is(err, svc.createErr) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("newDeps uses home dir", func(t *testing.T) {
		home := t.TempDir()
		original := os.Getenv("HOME")
		t.Setenv("HOME", home)
		t.Cleanup(func() { _ = os.Setenv("HOME", original) })

		dep, err := newDeps()
		if err != nil {
			t.Fatalf("newDeps() error = %v", err)
		}
		if dep.store.RaiseDir() != home+"/.raise" {
			t.Fatalf("RaiseDir() = %q", dep.store.RaiseDir())
		}
		if len(dep.registry) == 0 || dep.service == nil {
			t.Fatal("newDeps() returned incomplete dependencies")
		}
	})

	t.Run("newDeps returns error when home is unavailable", func(t *testing.T) {
		original := os.Getenv("HOME")
		t.Setenv("HOME", "")
		t.Cleanup(func() { _ = os.Setenv("HOME", original) })

		_, err := newDeps()
		if err == nil {
			t.Fatal("newDeps() error = nil")
		}
	})

	t.Run("active update helpers", func(t *testing.T) {
		store := &mockStore{active: map[string]string{"claude": "alpha"}}
		if err := renameActiveProfiles(store, "alpha", "beta"); err != nil {
			t.Fatalf("renameActiveProfiles() error = %v", err)
		}
		if err := removeActiveProfile(store, "beta"); err != nil {
			t.Fatalf("removeActiveProfile() error = %v", err)
		}
		if err := removeActiveProfile(store, "missing"); err != nil {
			t.Fatalf("removeActiveProfile(missing) error = %v", err)
		}
		if replaceActiveProfile(map[string]string{}, "a", "b") {
			t.Fatal("replaceActiveProfile() should report unchanged")
		}
		if dropActiveProfile(map[string]string{}, "a") {
			t.Fatal("dropActiveProfile() should report unchanged")
		}
	})

	t.Run("confirm delete accepts eof", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ok, err := confirmDelete(strings.NewReader("yes"), buf)
		if err != nil || !ok {
			t.Fatalf("confirmDelete() = %v, %v", ok, err)
		}
	})

	t.Run("confirm delete returns read error", func(t *testing.T) {
		buf := &bytes.Buffer{}
		_, err := confirmDelete(errReader{err: errors.New("read failed")}, buf)
		if err == nil || err.Error() != "read failed" {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("confirm delete returns write error", func(t *testing.T) {
		_, err := confirmDelete(strings.NewReader("y\n"), errWriter{err: errors.New("write failed")})
		if err == nil || err.Error() != "write failed" {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("remove active profile returns update error", func(t *testing.T) {
		store := &mockStore{
			active: map[string]string{"claude": "alpha"},
			setErr: errors.New("set active failed"),
		}
		if err := removeActiveProfile(store, "alpha"); !errors.Is(err, store.setErr) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("remove active profile returns lookup error", func(t *testing.T) {
		store := &mockStore{activeErr: errors.New("active lookup failed")}
		if err := removeActiveProfile(store, "alpha"); !errors.Is(err, store.activeErr) {
			t.Fatalf("error = %v", err)
		}
	})
}
