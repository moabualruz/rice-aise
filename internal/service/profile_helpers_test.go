package service

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/mkh/rice-aise/internal/domain"
)

func TestCredentialAndFilesystemHelpers(t *testing.T) {
	t.Run("copyCredentials handles relative absolute and dotfiles", func(t *testing.T) {
		root := t.TempDir()
		src := filepath.Join(root, "src")
		dst := filepath.Join(root, "dst")
		writeFile(t, filepath.Join(src, "std", "auth.json"), "std-auth")
		writeFile(t, filepath.Join(src, "multi-data", "auth.json"), "multi-auth")
		writeFile(t, filepath.Join(src, "aider", ".aider.env"), "aider-env")

		if err := copyCredentials(src, dst, standardTool()); err != nil {
			t.Fatalf("copyCredentials(std) error = %v", err)
		}
		if err := copyCredentials(src, dst, multiTool()); err != nil {
			t.Fatalf("copyCredentials(multi) error = %v", err)
		}
		if err := copyCredentials(src, dst, aiderTool()); err != nil {
			t.Fatalf("copyCredentials(aider) error = %v", err)
		}

		assertFile(t, filepath.Join(dst, "std", "auth.json"), "std-auth")
		assertFile(t, filepath.Join(dst, "multi-data", "auth.json"), "multi-auth")
		assertFile(t, filepath.Join(dst, "aider", ".aider.env"), "aider-env")
		if _, err := credentialPath(dst, domain.Tool{
			ID:              "bad",
			ConfigDirs:      []domain.DirMapping{{SourcePath: "~/.bad", ProfileSubdir: "bad"}},
			CredentialFiles: []string{"/tmp/outside"},
		}, "/tmp/outside"); err == nil {
			t.Fatal("credentialPath() error = nil, want non-nil")
		}
	})

	t.Run("copyCredentials returns destination credential path error", func(t *testing.T) {
		tool := domain.Tool{
			ID:              "bad",
			ConfigDirs:      []domain.DirMapping{{SourcePath: "/managed", ProfileSubdir: "bad"}},
			CredentialFiles: []string{"/managed/credential.json"},
		}
		original := credentialPathFunc
		credentialPathFunc = func(profileDir string, gotTool domain.Tool, credential string) (string, error) {
			if profileDir == "/dst" {
				return "", errors.New("credential \"/outside/credential.json\" outside managed paths")
			}
			return credentialPath(profileDir, gotTool, credential)
		}
		t.Cleanup(func() { credentialPathFunc = original })

		err := copyCredentials("/src", "/dst", tool)
		if err == nil || err.Error() != "credential \"/outside/credential.json\" outside managed paths" {
			t.Fatalf("copyCredentials() error = %v", err)
		}
	})

	t.Run("filesystem helpers cover file dir and symlink behavior", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "srcdir", "nested", "data.txt"), "payload")
		writeFile(t, filepath.Join(root, "srcfile.txt"), "file")

		if err := ensureProfileLayout(root, aiderTool()); err != nil {
			t.Fatalf("ensureProfileLayout() error = %v", err)
		}
		assertFile(t, filepath.Join(root, "aider", ".aider.conf.yml"), "")

		if err := moveOrSeedDir(filepath.Join(root, "srcdir"), filepath.Join(root, "moved")); err != nil {
			t.Fatalf("moveOrSeedDir(existing) error = %v", err)
		}
		if err := moveOrSeedDir(filepath.Join(root, "missing-dir"), filepath.Join(root, "seeded")); err != nil {
			t.Fatalf("moveOrSeedDir(missing) error = %v", err)
		}
		assertDir(t, filepath.Join(root, "seeded"))

		if err := moveOrSeedFile(filepath.Join(root, "srcfile.txt"), filepath.Join(root, "moved.txt")); err != nil {
			t.Fatalf("moveOrSeedFile(existing) error = %v", err)
		}
		if err := moveOrSeedFile(filepath.Join(root, "missing.txt"), filepath.Join(root, "seeded.txt")); err != nil {
			t.Fatalf("moveOrSeedFile(missing) error = %v", err)
		}
		assertFile(t, filepath.Join(root, "seeded.txt"), "")

		first := filepath.Join(root, "first")
		second := filepath.Join(root, "second")
		writeFile(t, first, "one")
		writeFile(t, second, "two")
		link := filepath.Join(root, "link")
		if err := os.Symlink(first, link); err != nil {
			t.Fatal(err)
		}
		if err := swapSymlink(second, link); err != nil {
			t.Fatalf("swapSymlink() error = %v", err)
		}
		assertLink(t, link, second)
		assertMissing(t, link+".tmp")

		resolved, err := resolveCopySource(link)
		if err != nil || resolved != second {
			t.Fatalf("resolveCopySource() = %q, %v", resolved, err)
		}
		if _, err := resolveCopySource(filepath.Join(root, "missing-link")); err == nil {
			t.Fatal("resolveCopySource() missing error = nil")
		}

		if err := copyEntry(filepath.Join(root, "moved"), filepath.Join(root, "copied-dir")); err != nil {
			t.Fatalf("copyEntry(dir) error = %v", err)
		}
		if err := copyEntry(second, filepath.Join(root, "copied.txt")); err != nil {
			t.Fatalf("copyEntry(file) error = %v", err)
		}
		assertFile(t, filepath.Join(root, "copied-dir", "nested", "data.txt"), "payload")
		assertFile(t, filepath.Join(root, "copied.txt"), "two")

		if err := copyCredentialFile(filepath.Join(root, "absent"), filepath.Join(root, "ignored")); err != nil {
			t.Fatalf("copyCredentialFile() error = %v", err)
		}
		if err := writeEmptyFile(filepath.Join(root, "empty", "file.txt")); err != nil {
			t.Fatalf("writeEmptyFile() error = %v", err)
		}
	})
}

func TestSimpleHelpers(t *testing.T) {
	tool := standardTool()
	ids := enabledToolIDs(&domain.Profile{Tools: map[string]bool{"b": false, "a": true}})
	slices.Sort(ids)
	if !reflect.DeepEqual(ids, []string{"a"}) {
		t.Fatalf("enabledToolIDs() = %v", ids)
	}
	if got := toolMap([]domain.Tool{tool}); !reflect.DeepEqual(got, map[string]bool{"std": true}) {
		t.Fatalf("toolMap() = %#v", got)
	}
	if got := activeMap([]domain.Tool{tool}, "alpha"); !reflect.DeepEqual(got, map[string]string{"std": "alpha"}) {
		t.Fatalf("activeMap() = %#v", got)
	}
	if got := profileConfigPath("/tmp/p", tool.ConfigDirs[0]); got != filepath.Join("/tmp/p", "std") {
		t.Fatalf("profileConfigPath() = %q", got)
	}
	if got := profileDotfilePath("/tmp/p", aiderTool(), ".aider.env"); got != filepath.Join("/tmp/p", "aider", ".aider.env") {
		t.Fatalf("profileDotfilePath() = %q", got)
	}
	if got := homePath("/tmp/home", ".aider.env"); got != filepath.Join("/tmp/home", ".aider.env") {
		t.Fatalf("homePath() = %q", got)
	}
}
