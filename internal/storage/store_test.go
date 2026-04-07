package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/mkh/rice-aise/internal/domain"
)

func TestFileProfileStorePathHelpers(t *testing.T) {
	store := NewFileProfileStore(t.TempDir())

	if got, want := store.RaiseDir(), store.raiseDir; got != want {
		t.Fatalf("RaiseDir() = %q, want %q", got, want)
	}

	wantDir := filepath.Join(store.RaiseDir(), "profiles", "alpha")
	if got := store.ProfileDir("alpha"); got != wantDir {
		t.Fatalf("ProfileDir() = %q, want %q", got, wantDir)
	}
}

func TestFileProfileStoreCreateAndGet(t *testing.T) {
	store := NewFileProfileStore(t.TempDir())
	tools := map[string]bool{"claude": true, "codex": false}

	profile, err := store.Create("alpha", tools)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if profile.Name != "alpha" {
		t.Fatalf("Create() name = %q, want alpha", profile.Name)
	}
	if profile.CreatedAt.IsZero() {
		t.Fatal("Create() returned zero CreatedAt")
	}
	if !reflect.DeepEqual(profile.Tools, tools) {
		t.Fatalf("Create() tools = %#v, want %#v", profile.Tools, tools)
	}

	onDisk := readMetadataSnapshot(t, filepath.Join(store.ProfileDir("alpha"), "metadata.json"))
	if onDisk.Name != "alpha" {
		t.Fatalf("metadata name = %q, want alpha", onDisk.Name)
	}
	if onDisk.CreatedAt != profile.CreatedAt.Format(time.RFC3339) {
		t.Fatalf("metadata createdAt = %q, want %q", onDisk.CreatedAt, profile.CreatedAt.Format(time.RFC3339))
	}
	if !reflect.DeepEqual(onDisk.Tools, tools) {
		t.Fatalf("metadata tools = %#v, want %#v", onDisk.Tools, tools)
	}

	got, err := store.Get("alpha")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !profilesEqual(*got, *profile) {
		t.Fatalf("Get() = %#v, want %#v", *got, *profile)
	}
}

func TestFileProfileStoreListSorted(t *testing.T) {
	store := NewFileProfileStore(t.TempDir())
	names := []string{"charlie", "alpha", "bravo"}

	for _, name := range names {
		if _, err := store.Create(name, map[string]bool{name: true}); err != nil {
			t.Fatalf("Create(%q) error = %v", name, err)
		}
	}

	list, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	gotNames := make([]string, 0, len(list))
	for _, profile := range list {
		gotNames = append(gotNames, profile.Name)
	}
	if !slices.Equal(gotNames, []string{"alpha", "bravo", "charlie"}) {
		t.Fatalf("List() names = %v, want %v", gotNames, []string{"alpha", "bravo", "charlie"})
	}
}

func TestFileProfileStoreListEmptyDir(t *testing.T) {
	store := NewFileProfileStore(t.TempDir())

	list, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("List() len = %d, want 0", len(list))
	}
}

func TestFileProfileStoreListInvalidJSON(t *testing.T) {
	store := NewFileProfileStore(t.TempDir())
	writeRawFile(t, filepath.Join(store.ProfileDir("broken"), "metadata.json"), "{")

	if _, err := store.List(); err == nil {
		t.Fatal("List() error = nil, want non-nil")
	}
}

func TestFileProfileStoreListReadDirError(t *testing.T) {
	store := NewFileProfileStore(t.TempDir())
	writeRawFile(t, filepath.Join(store.RaiseDir(), "profiles"), "file")

	if _, err := store.List(); err == nil {
		t.Fatal("List() error = nil, want non-nil")
	}
}

func TestFileProfileStoreDelete(t *testing.T) {
	store := NewFileProfileStore(t.TempDir())
	if _, err := store.Create("alpha", map[string]bool{"claude": true}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := store.Delete("alpha"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if _, err := os.Stat(store.ProfileDir("alpha")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("profile dir stat error = %v, want %v", err, os.ErrNotExist)
	}

	if _, err := store.Get("alpha"); err == nil {
		t.Fatal("Get() after Delete() error = nil, want non-nil")
	}
}

func TestFileProfileStoreRename(t *testing.T) {
	store := NewFileProfileStore(t.TempDir())
	if _, err := store.Create("old", map[string]bool{"codex": true}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := store.Rename("old", "new"); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}

	if _, err := os.Stat(store.ProfileDir("old")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("old profile dir stat error = %v, want %v", err, os.ErrNotExist)
	}

	profile, err := store.Get("new")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if profile.Name != "new" {
		t.Fatalf("renamed profile name = %q, want new", profile.Name)
	}
}

func TestFileProfileStoreRenameMissingProfile(t *testing.T) {
	store := NewFileProfileStore(t.TempDir())

	if err := store.Rename("missing", "new"); err == nil {
		t.Fatal("Rename() error = nil, want non-nil")
	}
}

func TestFileProfileStoreRenameBrokenMetadata(t *testing.T) {
	store := NewFileProfileStore(t.TempDir())
	writeRawFile(t, filepath.Join(store.ProfileDir("old"), "metadata.json"), `{"name":"old","createdAt":"bad","tools":{}}`)

	if err := store.Rename("old", "new"); err == nil {
		t.Fatal("Rename() error = nil, want non-nil")
	}
}

func TestFileProfileStoreSetAndGetActive(t *testing.T) {
	store := NewFileProfileStore(t.TempDir())
	active := map[string]string{"claude": "alpha", "codex": "beta"}

	if err := store.SetActive(active); err != nil {
		t.Fatalf("SetActive() error = %v", err)
	}

	config := readConfigSnapshot(t, filepath.Join(store.RaiseDir(), "config.json"))
	if !reflect.DeepEqual(config.Active, active) {
		t.Fatalf("config active = %#v, want %#v", config.Active, active)
	}

	got, err := store.GetActive()
	if err != nil {
		t.Fatalf("GetActive() error = %v", err)
	}
	if !reflect.DeepEqual(got, active) {
		t.Fatalf("GetActive() = %#v, want %#v", got, active)
	}
}

func TestFileProfileStoreGetActiveEmptyDir(t *testing.T) {
	store := NewFileProfileStore(t.TempDir())

	active, err := store.GetActive()
	if err != nil {
		t.Fatalf("GetActive() error = %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("GetActive() len = %d, want 0", len(active))
	}
}

func TestFileProfileStoreGetNonExistentProfile(t *testing.T) {
	store := NewFileProfileStore(t.TempDir())

	if _, err := store.Get("missing"); err == nil {
		t.Fatal("Get() error = nil, want non-nil")
	}
}

func TestFileProfileStoreGetInvalidJSON(t *testing.T) {
	store := NewFileProfileStore(t.TempDir())
	writeRawFile(t, filepath.Join(store.ProfileDir("broken"), "metadata.json"), "{")

	if _, err := store.Get("broken"); err == nil {
		t.Fatal("Get() error = nil, want non-nil")
	}
}

func TestFileProfileStoreGetInvalidCreatedAt(t *testing.T) {
	store := NewFileProfileStore(t.TempDir())
	writeRawFile(t, filepath.Join(store.ProfileDir("broken"), "metadata.json"), `{"name":"broken","createdAt":"bad","tools":{}}`)

	if _, err := store.Get("broken"); err == nil {
		t.Fatal("Get() error = nil, want non-nil")
	}
}

func TestFileProfileStoreGetActiveInvalidJSON(t *testing.T) {
	store := NewFileProfileStore(t.TempDir())
	writeRawFile(t, filepath.Join(store.RaiseDir(), "config.json"), "{")

	if _, err := store.GetActive(); err == nil {
		t.Fatal("GetActive() error = nil, want non-nil")
	}
}

func TestFileProfileStoreCreateError(t *testing.T) {
	blocked := filepath.Join(t.TempDir(), "blocked")
	writeRawFile(t, blocked, "file")
	store := NewFileProfileStore(blocked)

	if _, err := store.Create("alpha", map[string]bool{"claude": true}); err == nil {
		t.Fatal("Create() error = nil, want non-nil")
	}
}

func TestFileProfileStoreWriteJSONMarshalError(t *testing.T) {
	store := NewFileProfileStore(t.TempDir())

	if err := store.writeJSON(filepath.Join(store.RaiseDir(), "bad.json"), func() {}); err == nil {
		t.Fatal("writeJSON() error = nil, want non-nil")
	}
}

func profilesEqual(a, b domain.Profile) bool {
	return a.Name == b.Name &&
		a.CreatedAt.Equal(b.CreatedAt) &&
		reflect.DeepEqual(a.Tools, b.Tools)
}

type metadataSnapshot struct {
	Name      string          `json:"name"`
	CreatedAt string          `json:"createdAt"`
	Tools     map[string]bool `json:"tools"`
}

type configSnapshot struct {
	Active map[string]string `json:"active"`
}

func readMetadataSnapshot(t *testing.T, path string) metadataSnapshot {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	var meta metadataSnapshot
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("Unmarshal(%q) error = %v", path, err)
	}
	if _, err := time.Parse(time.RFC3339, meta.CreatedAt); err != nil {
		t.Fatalf("createdAt parse error = %v", err)
	}
	return meta
}

func readConfigSnapshot(t *testing.T, path string) configSnapshot {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	var cfg configSnapshot
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("Unmarshal(%q) error = %v", path, err)
	}
	return cfg
}

func writeRawFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
