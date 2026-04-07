package service

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mkh/rice-aise/internal/domain"
)

func TestNewProfileServiceAndDetectInstalled(t *testing.T) {
	home := t.TempDir()
	resolver := &mockResolver{homeDir: home, exists: map[string]bool{
		filepath.Join(home, ".std"): true,
	}}
	registry := []domain.Tool{
		standardTool(),
		{ID: "ghost"},
		multiTool(),
	}
	service := NewProfileService(newMockStore(t.TempDir()), resolver, registry)

	if service.store == nil || service.resolver == nil || len(service.registry) != 3 {
		t.Fatal("constructor did not populate service")
	}

	got := idsOf(service.DetectInstalled())
	if !reflect.DeepEqual(got, []string{"std"}) {
		t.Fatalf("DetectInstalled() = %v, want [std]", got)
	}
}

func TestInitMovesCurrentConfigsAndCreatesVanilla(t *testing.T) {
	home, _, svc, store := newHarness(t)
	writeLiveStandard(t, home)
	writeLiveAider(t, home)
	writeLiveMulti(t, home)
	restore := setNow(t, time.Date(2026, 4, 7, 9, 0, 0, 0, time.UTC))
	defer restore()

	current, err := svc.Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if current != "2026-04-07-current" {
		t.Fatalf("Init() current = %q", current)
	}

	assertLink(t, filepath.Join(home, ".std"), filepath.Join(store.ProfileDir(current), "std"))
	assertLink(t, filepath.Join(home, ".aider"), filepath.Join(store.ProfileDir(current), "aider-cache"))
	assertLink(t, filepath.Join(home, ".aider.conf.yml"), filepath.Join(store.ProfileDir(current), "aider", ".aider.conf.yml"))
	assertLink(t, filepath.Join(home, ".aider.env"), filepath.Join(store.ProfileDir(current), "aider", ".aider.env"))
	assertLink(t, filepath.Join(home, ".config", "multi"), filepath.Join(store.ProfileDir(current), "multi-config"))
	assertLink(t, filepath.Join(home, ".local", "share", "multi"), filepath.Join(store.ProfileDir(current), "multi-data"))

	assertFile(t, filepath.Join(store.ProfileDir(current), "std", "config.txt"), "std-config")
	assertFile(t, filepath.Join(store.ProfileDir(current), "std", "auth.json"), "std-auth")
	assertFile(t, filepath.Join(store.ProfileDir(current), "aider-cache", "cache.txt"), "aider-cache")
	assertFile(t, filepath.Join(store.ProfileDir(current), "aider", ".aider.conf.yml"), "aider-conf")
	assertFile(t, filepath.Join(store.ProfileDir(current), "aider", ".aider.env"), "aider-env")
	assertFile(t, filepath.Join(store.ProfileDir(current), "multi-config", "settings.yml"), "multi-config")
	assertFile(t, filepath.Join(store.ProfileDir(current), "multi-data", "auth.json"), "multi-auth")
	assertFile(t, filepath.Join(store.ProfileDir(current), "multi-data", "data.txt"), "multi-data")

	assertFile(t, filepath.Join(store.ProfileDir(vanillaProfileName), "std", "auth.json"), "std-auth")
	assertMissing(t, filepath.Join(store.ProfileDir(vanillaProfileName), "std", "config.txt"))
	assertFile(t, filepath.Join(store.ProfileDir(vanillaProfileName), "aider", ".aider.env"), "aider-env")
	assertFile(t, filepath.Join(store.ProfileDir(vanillaProfileName), "aider", ".aider.conf.yml"), "")
	assertDir(t, filepath.Join(store.ProfileDir(vanillaProfileName), "multi-config"))
	assertFile(t, filepath.Join(store.ProfileDir(vanillaProfileName), "multi-data", "auth.json"), "multi-auth")
	assertMissing(t, filepath.Join(store.ProfileDir(vanillaProfileName), "multi-data", "data.txt"))

	wantActive := map[string]string{"std": current, "aider": current, "multi": current}
	if !reflect.DeepEqual(store.active, wantActive) {
		t.Fatalf("active mapping = %#v, want %#v", store.active, wantActive)
	}
	if _, ok := store.profiles[vanillaProfileName]; !ok {
		t.Fatal("vanilla profile was not created")
	}
}

func TestInitErrors(t *testing.T) {
	t.Run("already initialized", func(t *testing.T) {
		_, _, svc, store := newHarness(t)
		store.profiles["existing"] = &domain.Profile{Name: "existing", Tools: map[string]bool{"std": true}}
		if _, err := svc.Init(); err == nil || !strings.Contains(err.Error(), "already initialized") {
			t.Fatalf("Init() error = %v, want already initialized", err)
		}
	})

	t.Run("refuses symlinked live dir", func(t *testing.T) {
		home, _, svc, _ := newHarness(t)
		target := filepath.Join(t.TempDir(), "target")
		writeFile(t, filepath.Join(target, "config.txt"), "x")
		if err := os.MkdirAll(filepath.Dir(filepath.Join(home, ".std")), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filepath.Join(home, ".std")); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.Init(); err == nil || !strings.Contains(err.Error(), "already linked") {
			t.Fatalf("Init() error = %v, want already linked", err)
		}
	})
}

func TestSaveCopiesLiveConfigFollowingSymlinks(t *testing.T) {
	home, _, svc, store := newHarness(t)
	writeLiveStandard(t, home)
	writeLiveAider(t, home)
	writeLiveMulti(t, home)
	restore := setNow(t, time.Date(2026, 4, 7, 9, 0, 0, 0, time.UTC))
	defer restore()
	current, err := svc.Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	writeFile(t, filepath.Join(store.ProfileDir(current), "std", "extra.txt"), "extra")
	writeFile(t, filepath.Join(store.ProfileDir(current), "multi-config", "nested", "more.txt"), "nested")
	if err := os.Remove(filepath.Join(home, ".aider.conf.yml")); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	if err := svc.Save("snapshot"); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	assertFile(t, filepath.Join(store.ProfileDir("snapshot"), "std", "config.txt"), "std-config")
	assertFile(t, filepath.Join(store.ProfileDir("snapshot"), "std", "extra.txt"), "extra")
	assertFile(t, filepath.Join(store.ProfileDir("snapshot"), "multi-config", "nested", "more.txt"), "nested")
	assertFile(t, filepath.Join(store.ProfileDir("snapshot"), "multi-data", "auth.json"), "multi-auth")
	assertFile(t, filepath.Join(store.ProfileDir("snapshot"), "aider", ".aider.conf.yml"), "")
	assertFile(t, filepath.Join(store.ProfileDir("snapshot"), "aider", ".aider.env"), "aider-env")
}

func TestSaveUnknownActiveTool(t *testing.T) {
	_, _, svc, store := newHarness(t)
	store.active["ghost"] = "alpha"
	if err := svc.Save("snapshot"); err == nil || !strings.Contains(err.Error(), "unknown tool") {
		t.Fatalf("Save() error = %v, want unknown tool", err)
	}
}

func TestUseSwitchesProfilesAndSupportsSubset(t *testing.T) {
	home, _, svc, store := newHarness(t)
	createProfileData(t, store.ProfileDir("alpha"), "alpha")
	createProfileData(t, store.ProfileDir("beta"), "beta")
	store.profiles["alpha"] = &domain.Profile{Name: "alpha", Tools: map[string]bool{"std": true, "aider": true, "multi": true}}
	store.profiles["beta"] = &domain.Profile{Name: "beta", Tools: map[string]bool{"std": true, "aider": true, "multi": true}}
	store.active = map[string]string{"std": "alpha", "aider": "alpha", "multi": "alpha"}

	linkProfileToHome(t, home, store.ProfileDir("alpha"))
	if err := svc.Use("beta", []string{"std"}); err != nil {
		t.Fatalf("Use(subset) error = %v", err)
	}
	assertLink(t, filepath.Join(home, ".std"), filepath.Join(store.ProfileDir("beta"), "std"))
	assertLink(t, filepath.Join(home, ".aider"), filepath.Join(store.ProfileDir("alpha"), "aider-cache"))
	if got := store.active["aider"]; got != "alpha" {
		t.Fatalf("aider active = %q, want alpha", got)
	}

	if err := svc.Use("beta", nil); err != nil {
		t.Fatalf("Use(all) error = %v", err)
	}
	assertLink(t, filepath.Join(home, ".aider"), filepath.Join(store.ProfileDir("beta"), "aider-cache"))
	assertLink(t, filepath.Join(home, ".aider.env"), filepath.Join(store.ProfileDir("beta"), "aider", ".aider.env"))
	assertLink(t, filepath.Join(home, ".config", "multi"), filepath.Join(store.ProfileDir("beta"), "multi-config"))
	if got := store.active["multi"]; got != "beta" {
		t.Fatalf("multi active = %q, want beta", got)
	}
}

func TestUseErrors(t *testing.T) {
	t.Run("tool not in profile", func(t *testing.T) {
		_, _, svc, store := newHarness(t)
		store.profiles["beta"] = &domain.Profile{Name: "beta", Tools: map[string]bool{"std": true}}
		if err := svc.Use("beta", []string{"aider"}); err == nil || !strings.Contains(err.Error(), "does not include") {
			t.Fatalf("Use() error = %v, want profile mismatch", err)
		}
	})

	t.Run("unknown tool in profile", func(t *testing.T) {
		_, _, svc, store := newHarness(t)
		store.profiles["beta"] = &domain.Profile{Name: "beta", Tools: map[string]bool{"ghost": true}}
		if err := svc.Use("beta", nil); err == nil || !strings.Contains(err.Error(), "unknown tool") {
			t.Fatalf("Use() error = %v, want unknown tool", err)
		}
	})
}

func TestCreateCopiesOnlyCredentials(t *testing.T) {
	home, _, svc, store := newHarness(t)
	writeLiveStandard(t, home)
	writeLiveAider(t, home)
	writeLiveMulti(t, home)
	restore := setNow(t, time.Date(2026, 4, 7, 9, 0, 0, 0, time.UTC))
	defer restore()
	current, err := svc.Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	delete(store.active, "multi")

	if err := svc.Create("fresh"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	assertFile(t, filepath.Join(store.ProfileDir("fresh"), "std", "auth.json"), "std-auth")
	assertMissing(t, filepath.Join(store.ProfileDir("fresh"), "std", "config.txt"))
	assertFile(t, filepath.Join(store.ProfileDir("fresh"), "aider", ".aider.env"), "aider-env")
	assertFile(t, filepath.Join(store.ProfileDir("fresh"), "aider", ".aider.conf.yml"), "")
	assertDir(t, filepath.Join(store.ProfileDir("fresh"), "multi-config"))
	assertDir(t, filepath.Join(store.ProfileDir("fresh"), "multi-data"))
	assertMissing(t, filepath.Join(store.ProfileDir("fresh"), "multi-data", "auth.json"))
	if got := store.active["std"]; got != current {
		t.Fatalf("Create() should not update active mapping, got %q", got)
	}
}

func TestWhich(t *testing.T) {
	_, _, svc, store := newHarness(t)
	store.active["std"] = "alpha"
	got, err := svc.Which("std")
	if err != nil || got != "alpha" {
		t.Fatalf("Which() = %q, %v", got, err)
	}
	if _, err := svc.Which("missing"); err == nil {
		t.Fatal("Which() error = nil, want non-nil")
	}
	store.getActiveErr = errors.New("boom")
	if _, err := svc.Which("std"); !errors.Is(err, store.getActiveErr) {
		t.Fatalf("Which() error = %v, want %v", err, store.getActiveErr)
	}
}
