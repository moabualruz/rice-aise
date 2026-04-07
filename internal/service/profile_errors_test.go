package service

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mkh/rice-aise/internal/domain"
)

func TestErrorPaths(t *testing.T) {
	t.Run("service methods propagate store and fs errors", func(t *testing.T) {
		home, _, svc, store := newHarness(t)
		writeLiveStandard(t, home)
		writeLiveAider(t, home)
		writeLiveMulti(t, home)

		store.listErr = errors.New("list")
		if err := svc.ensureUninitialized(); !errors.Is(err, store.listErr) {
			t.Fatalf("ensureUninitialized() error = %v", err)
		}
		store.listErr = nil

		store.createErr = errors.New("create")
		if _, err := svc.Init(); !errors.Is(err, store.createErr) {
			t.Fatalf("Init(create) error = %v", err)
		}
		store.createErr = nil
		restore := setNow(t, time.Date(2026, 4, 7, 9, 0, 0, 0, time.UTC))
		defer restore()
		if _, err := svc.Init(); err != nil {
			t.Fatalf("Init() error = %v", err)
		}
		store.createErr = errors.New("create")
		if err := svc.Save("bad-save"); !errors.Is(err, store.createErr) {
			t.Fatalf("Save(create) error = %v", err)
		}
		store.getActiveErr = errors.New("active")
		if err := svc.Save("bad-active"); !errors.Is(err, store.getActiveErr) {
			t.Fatalf("Save(active) error = %v", err)
		}
		if err := svc.Create("bad-create"); !errors.Is(err, store.getActiveErr) {
			t.Fatalf("Create(active) error = %v", err)
		}
		store.getActiveErr = nil
		store.createErr = errors.New("create")
		if err := svc.Create("bad-create"); !errors.Is(err, store.createErr) {
			t.Fatalf("Create(create) error = %v", err)
		}
		store.createErr = nil
		store.getActiveErr = errors.New("active")
		if err := svc.Use("missing", nil); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("Use(get) error = %v", err)
		}
		store.profiles["alpha"] = &domain.Profile{Name: "alpha", Tools: map[string]bool{"std": true}}
		if err := svc.Use("alpha", nil); !errors.Is(err, store.getActiveErr) {
			t.Fatalf("Use(active) error = %v", err)
		}
	})

	t.Run("init fails when vanilla creation fails", func(t *testing.T) {
		home, _, svc, store := newHarness(t)
		writeLiveStandard(t, home)
		writeLiveAider(t, home)
		writeLiveMulti(t, home)
		restore := setNow(t, time.Date(2026, 4, 7, 9, 0, 0, 0, time.UTC))
		defer restore()
		store.createErr = errors.New("vanilla")
		store.failCreateAt = 2
		if _, err := svc.Init(); !errors.Is(err, store.createErr) {
			t.Fatalf("Init(vanilla) error = %v", err)
		}
	})

	t.Run("use reports swap failures", func(t *testing.T) {
		home, _, svc, store := newHarness(t)
		store.profiles["beta"] = &domain.Profile{Name: "beta", Tools: map[string]bool{"std": true}}
		store.active["std"] = "alpha"
		createProfileData(t, store.ProfileDir("alpha"), "alpha")
		linkProfileToHome(t, home, store.ProfileDir("alpha"))
		writeFile(t, filepath.Join(home, ".std.tmp", "block"), "x")
		if err := svc.Use("beta", []string{"std"}); err == nil {
			t.Fatal("Use(swap) error = nil")
		}
	})

	t.Run("helper methods report fs errors", func(t *testing.T) {
		home, _, svc, store := newHarness(t)
		blockedProfile := store.ProfileDir("blocked")
		writeFile(t, blockedProfile, "file")
		writeLiveStandard(t, home)
		if err := svc.captureCurrentTool("blocked", standardTool()); err == nil {
			t.Fatal("captureCurrentTool() error = nil")
		}

		writeFile(t, filepath.Join(home, ".std.tmp", "block"), "x")
		if err := svc.moveCurrentDirs("swapfail", standardTool()); err == nil {
			t.Fatal("moveCurrentDirs() error = nil")
		}

		writeLiveAider(t, home)
		writeFile(t, filepath.Join(store.ProfileDir("dotfile-block"), "aider"), "file")
		if err := svc.moveCurrentDotfiles("dotfile-block", aiderTool()); err == nil {
			t.Fatal("moveCurrentDotfiles() error = nil")
		}

		writeLiveAider(t, home)
		writeFile(t, filepath.Join(home, ".aider.conf.yml.tmp", "block"), "x")
		if err := svc.moveCurrentDotfiles("dotfile-swap", aiderTool()); err == nil {
			t.Fatal("moveCurrentDotfiles(swap) error = nil")
		}

		store.createErr = errors.New("create")
		if err := svc.createVanillaProfile("current", []domain.Tool{standardTool()}); !errors.Is(err, store.createErr) {
			t.Fatalf("createVanillaProfile() error = %v", err)
		}

		writeFile(t, store.ProfileDir("bad-layout"), "file")
		if err := svc.seedProfile("bad-layout", []domain.Tool{standardTool()}, map[string]string{}); err == nil {
			t.Fatal("seedProfile(layout) error = nil")
		}
		badTool := domain.Tool{ID: "bad", ConfigDirs: []domain.DirMapping{{SourcePath: "~/.bad", ProfileSubdir: "bad"}}, CredentialFiles: []string{"/tmp/outside"}}
		if err := svc.seedProfile("bad-cred", []domain.Tool{badTool}, map[string]string{"bad": "src"}); err == nil {
			t.Fatal("seedProfile(credentials) error = nil")
		}

		if err := svc.copyLiveTools(store.ProfileDir("copy"), []domain.Tool{standardTool()}); err == nil {
			t.Fatal("copyLiveTools() error = nil")
		}
		writeFile(t, filepath.Join(home, ".std"), "file")
		if err := svc.copyLiveTool(blockedProfile, standardTool()); err == nil {
			t.Fatal("copyLiveTool(layout) error = nil")
		}
		writeFile(t, filepath.Join(home, ".std"), "file")
		if err := svc.copyLiveTool(store.ProfileDir("copy2"), standardTool()); err == nil {
			t.Fatal("copyLiveTool(copy) error = nil")
		}

		if err := os.RemoveAll(filepath.Join(home, ".std")); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(home, "missing-target"), filepath.Join(home, ".std")); err != nil {
			t.Fatal(err)
		}
		if err := svc.copyLiveDirs(store.ProfileDir("copy3"), standardTool()); err == nil {
			t.Fatal("copyLiveDirs() error = nil")
		}
		badDotTool := aiderTool()
		badDotTool.Dotfiles = []string{"bad\x00name"}
		if err := svc.copyLiveDotfiles(store.ProfileDir("copy4"), badDotTool); err == nil {
			t.Fatal("copyLiveDotfiles(resolve) error = nil")
		}
		if err := os.RemoveAll(filepath.Join(home, ".aider.env")); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(home, "missing-target"), filepath.Join(home, ".aider.env")); err != nil {
			t.Fatal(err)
		}
		if err := svc.copyLiveDotfiles(store.ProfileDir("copy5"), aiderTool()); err == nil {
			t.Fatal("copyLiveDotfiles(copy) error = nil")
		}

		for _, path := range []string{
			filepath.Join(home, ".std"),
			filepath.Join(home, ".aider"),
			filepath.Join(home, ".aider.conf.yml"),
			filepath.Join(home, ".aider.env"),
			filepath.Join(home, ".config", "multi"),
			filepath.Join(home, ".local", "share", "multi"),
		} {
			if err := os.RemoveAll(path); err != nil {
				t.Fatal(err)
			}
		}
		linkProfileToHome(t, home, store.ProfileDir("alpha"))
		if err := os.MkdirAll(filepath.Join(home, ".std.tmp"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := svc.linkProfileTools("alpha", []domain.Tool{standardTool()}); err == nil {
			t.Fatal("linkProfileTools() error = nil")
		}
		if err := svc.linkProfileDotfiles("alpha", aiderTool()); err == nil {
			t.Fatal("linkProfileDotfiles() error = nil")
		}
	})

	t.Run("low-level helpers report edge-case errors", func(t *testing.T) {
		root := t.TempDir()
		block := filepath.Join(root, "block")
		writeFile(t, block, "file")
		if err := ensureProfileLayout(block, standardTool()); err == nil {
			t.Fatal("ensureProfileLayout() error = nil")
		}
		if err := ensureDotfileLayout(block, aiderTool()); err == nil {
			t.Fatal("ensureDotfileLayout() error = nil")
		}
		if err := moveOrSeedDir(filepath.Join(root, "src"), filepath.Join(block, "dir")); err == nil {
			t.Fatal("moveOrSeedDir() error = nil")
		}
		writeFile(t, filepath.Join(root, "src.txt"), "x")
		if err := moveOrSeedFile(filepath.Join(root, "src.txt"), filepath.Join(block, "file.txt")); err == nil {
			t.Fatal("moveOrSeedFile() error = nil")
		}
		if err := swapSymlink(filepath.Join(root, "target"), filepath.Join(block, "link")); err == nil {
			t.Fatal("swapSymlink(mkdir) error = nil")
		}
		writeFile(t, filepath.Join(root, "link.tmp", "block"), "x")
		if err := swapSymlink(filepath.Join(root, "target"), filepath.Join(root, "link")); err == nil {
			t.Fatal("swapSymlink(symlink) error = nil")
		}

		broken := filepath.Join(root, "broken")
		if err := os.Symlink(filepath.Join(root, "missing"), broken); err != nil {
			t.Fatal(err)
		}
		if got, err := resolveCopySource(broken); err != nil || got != broken {
			t.Fatalf("resolveCopySource(broken) = %q, %v", got, err)
		}
		if err := copyEntry(filepath.Join(root, "missing"), filepath.Join(root, "dst")); err == nil {
			t.Fatal("copyEntry() error = nil")
		}

		src := filepath.Join(root, "srcdir")
		writeFile(t, filepath.Join(src, "ok.txt"), "ok")
		locked := filepath.Join(src, "locked")
		if err := os.MkdirAll(locked, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(locked, "bad.txt"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(locked, 0); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(locked, 0o755)
		if err := copyDir(src, filepath.Join(root, "copy")); err == nil {
			t.Fatal("copyDir(walk) error = nil")
		}

		roParent := filepath.Join(root, "ro")
		dst := filepath.Join(roParent, "dst")
		if err := os.MkdirAll(dst, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(roParent, 0o555); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(roParent, 0o755)
		if err := copyDir(src, dst); err == nil {
			t.Fatal("copyDir(removeall) error = nil")
		}

		if err := copyFile(filepath.Join(root, "missing"), filepath.Join(root, "dst.txt")); err == nil {
			t.Fatal("copyFile(open) error = nil")
		}
		noRead := filepath.Join(root, "no-read.txt")
		writeFile(t, noRead, "secret")
		if err := os.Chmod(noRead, 0); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(noRead, 0o644)
		if err := copyFile(noRead, filepath.Join(root, "dst-no-read.txt")); err == nil {
			t.Fatal("copyFile(stat-then-open) error = nil")
		}
		if err := copyFile(filepath.Join(src, "ok.txt"), filepath.Join(block, "dst.txt")); err == nil {
			t.Fatal("copyFile(mkdir) error = nil")
		}
		if err := copyFile(filepath.Join(src, "ok.txt"), src); err == nil {
			t.Fatal("copyFile(openfile) error = nil")
		}
		if err := writeEmptyFile(filepath.Join(block, "empty.txt")); err == nil {
			t.Fatal("writeEmptyFile() error = nil")
		}
		dirCred := filepath.Join(root, "cred-dir")
		if err := os.MkdirAll(dirCred, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := copyCredentials(root, filepath.Join(root, "dst"), domain.Tool{
			ID:              "dircred",
			ConfigDirs:      []domain.DirMapping{{SourcePath: "~/.cred", ProfileSubdir: "cred-dir"}},
			CredentialFiles: []string{"."},
		}); err == nil {
			t.Fatal("copyCredentials(filecopy) error = nil")
		}
	})
}
