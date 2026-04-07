package cli

import (
	"bytes"
	"errors"
	"testing"

	"github.com/spf13/cobra"
)

func TestDeleteCommandErrorPaths(t *testing.T) {
	t.Run("delete propagates store error", func(t *testing.T) {
		t.Parallel()

		dep, _, store := newTestDeps()
		store.deleteErr = errors.New("delete failed")

		_, err := executeCommand(dep, "", "delete", "beta", "--force")
		if !errors.Is(err, store.deleteErr) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("delete propagates active profile update error", func(t *testing.T) {
		t.Parallel()

		dep, _, store := newTestDeps()
		store.setErr = errors.New("set active failed")

		_, err := executeCommand(dep, "", "delete", "beta", "--force")
		if !errors.Is(err, store.setErr) {
			t.Fatalf("error = %v", err)
		}
		if store.deleted != "beta" {
			t.Fatalf("deleted = %q", store.deleted)
		}
	})
}

func TestConfirmDeleteAnswers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "yes", input: "y\n", want: true},
		{name: "no", input: "n\n", want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			ok, err := confirmDelete(bytes.NewBufferString(tc.input), buf)
			if err != nil {
				t.Fatalf("confirmDelete() error = %v", err)
			}
			if ok != tc.want {
				t.Fatalf("confirmDelete() = %v, want %v", ok, tc.want)
			}
			if got := buf.String(); got != "Are you sure? [y/N]: " {
				t.Fatalf("prompt = %q", got)
			}
		})
	}
}

func TestFinishDeletePromptCancelled(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)

	if err := finishDeletePrompt(cmd, nil, false); err != nil {
		t.Fatalf("finishDeletePrompt() error = %v", err)
	}
	if got := buf.String(); got != "Delete cancelled.\n" {
		t.Fatalf("output = %q", got)
	}
}

func TestFinishDeletePromptOtherBranches(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	if err := finishDeletePrompt(cmd, errors.New("prompt failed"), false); err == nil || err.Error() != "prompt failed" {
		t.Fatalf("error = %v", err)
	}
	if err := finishDeletePrompt(cmd, nil, true); err != nil {
		t.Fatalf("finishDeletePrompt(ok) error = %v", err)
	}
}
