package cli

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestExitWithError(t *testing.T) {
	var code int

	originalExit := exitFunc
	exitFunc = func(c int) { code = c }
	t.Cleanup(func() { exitFunc = originalExit })

	buf := &bytes.Buffer{}
	exitWithError(buf, errors.New("boom"))

	if code != 1 {
		t.Fatalf("exit code = %d", code)
	}
	if got := buf.String(); got != "boom\n" {
		t.Fatalf("output = %q", got)
	}
}

func TestExecute(t *testing.T) {
	originalDepsFactory := depsFactory
	originalExit := exitFunc
	originalStdin := stdin
	originalStdout := stdout
	originalStderr := stderr
	originalArgs := os.Args
	t.Cleanup(func() {
		depsFactory = originalDepsFactory
		exitFunc = originalExit
		stdin = originalStdin
		stdout = originalStdout
		stderr = originalStderr
		os.Args = originalArgs
	})

	t.Run("execute runs command successfully", func(t *testing.T) {
		dep, _, _ := newTestDeps()
		out := &bytes.Buffer{}
		errOut := &bytes.Buffer{}

		depsFactory = func() (dependencies, error) { return dep, nil }
		exitFunc = func(int) {}
		stdin = strings.NewReader("")
		stdout = out
		stderr = errOut
		os.Args = []string{"raise", "version"}

		Execute()

		if !strings.Contains(out.String(), "raise v0.1.0") {
			t.Fatalf("stdout = %q", out.String())
		}
		if errOut.Len() != 0 {
			t.Fatalf("stderr = %q", errOut.String())
		}
	})

	t.Run("execute reports dependency construction error", func(t *testing.T) {
		out := &bytes.Buffer{}
		errOut := &bytes.Buffer{}
		var exitCode int

		depsFactory = func() (dependencies, error) { return dependencies{}, errors.New("deps failed") }
		exitFunc = func(code int) { exitCode = code }
		stdin = strings.NewReader("")
		stdout = out
		stderr = errOut
		os.Args = []string{"raise", "version"}

		Execute()

		if exitCode != 1 {
			t.Fatalf("exit code = %d", exitCode)
		}
		if !strings.Contains(errOut.String(), "deps failed") {
			t.Fatalf("stderr = %q", errOut.String())
		}
	})

	t.Run("execute reports command error", func(t *testing.T) {
		dep, _, _ := newTestDeps()
		out := &bytes.Buffer{}
		errOut := &bytes.Buffer{}
		var exitCode int

		depsFactory = func() (dependencies, error) { return dep, nil }
		exitFunc = func(code int) { exitCode = code }
		stdin = strings.NewReader("")
		stdout = out
		stderr = errOut
		os.Args = []string{"raise", "use"}

		Execute()

		if exitCode != 1 {
			t.Fatalf("exit code = %d", exitCode)
		}
		if !strings.Contains(errOut.String(), "accepts 1 arg(s), received 0") {
			t.Fatalf("stderr = %q", errOut.String())
		}
	})
}
