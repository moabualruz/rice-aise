package cli

import (
	"bytes"
	"io"
	"strings"

	"github.com/mkh/rice-aise/internal/domain"
)

// Execute() still ends in exitWithError() for fatal paths, and newDeps()'s
// os.UserHomeDir failure path is not injectable without changing production code.
func executeCommand(dep dependencies, input string, args ...string) (string, error) {
	buf := &bytes.Buffer{}
	cmd := newRootCommand(dep)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetIn(strings.NewReader(input))
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func newTestDeps() (dependencies, *mockService, *mockStore) {
	registry := []domain.Tool{
		{
			ID:              "claude",
			Name:            "Claude Code",
			ConfigDirs:      []domain.DirMapping{{SourcePath: "~/.claude", ProfileSubdir: "claude"}},
			CredentialFiles: []string{".credentials.json"},
		},
		{
			ID:         "codex",
			Name:       "Codex CLI",
			ConfigDirs: []domain.DirMapping{{SourcePath: "~/.codex", ProfileSubdir: "codex"}},
		},
	}
	svc := &mockService{initName: "current", whichName: "alpha"}
	store := &mockStore{
		raiseDir: "/tmp/.raise",
		profiles: []domain.Profile{{Name: "beta"}, {Name: "alpha"}},
		active:   map[string]string{"claude": "alpha", "codex": "beta"},
	}
	return dependencies{service: svc, store: store, registry: registry}, svc, store
}

type mockService struct {
	initName  string
	whichName string
	useName   string
	useOnly   []string
	initErr   error
	saveErr   error
	useErr    error
	createErr error
	whichErr  error
}

func (m *mockService) DetectInstalled() []domain.Tool { return nil }

func (m *mockService) Init() (string, error) {
	return m.initName, m.initErr
}

func (m *mockService) Save(name string) error {
	_ = name
	return m.saveErr
}

func (m *mockService) Use(profileName string, onlyTools []string) error {
	m.useName = profileName
	m.useOnly = append([]string(nil), onlyTools...)
	return m.useErr
}

func (m *mockService) Create(name string) error {
	_ = name
	return m.createErr
}

func (m *mockService) Which(toolID string) (string, error) {
	_ = toolID
	return m.whichName, m.whichErr
}

type mockStore struct {
	raiseDir  string
	profiles  []domain.Profile
	active    map[string]string
	renameOld string
	renameNew string
	deleted   string
	listErr   error
	activeErr error
	renameErr error
	deleteErr error
	setErr    error
}

func (m *mockStore) List() ([]domain.Profile, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return append([]domain.Profile(nil), m.profiles...), nil
}

func (m *mockStore) Delete(name string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deleted = name
	return nil
}

func (m *mockStore) Rename(oldName, newName string) error {
	if m.renameErr != nil {
		return m.renameErr
	}
	m.renameOld = oldName
	m.renameNew = newName
	return nil
}

func (m *mockStore) SetActive(mapping map[string]string) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.active = cloneStringMap(mapping)
	return nil
}

func (m *mockStore) GetActive() (map[string]string, error) {
	if m.activeErr != nil {
		return nil, m.activeErr
	}
	return cloneStringMap(m.active), nil
}

func (m *mockStore) RaiseDir() string {
	return m.raiseDir
}

func cloneStringMap(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

type errWriter struct {
	err error
}

func (w errWriter) Write(_ []byte) (int, error) {
	if w.err == nil {
		w.err = io.ErrClosedPipe
	}
	return 0, w.err
}

type errReader struct {
	err error
}

func (r errReader) Read(_ []byte) (int, error) {
	if r.err == nil {
		r.err = io.ErrUnexpectedEOF
	}
	return 0, r.err
}

type failAfterWriter struct {
	writes int
	failOn int
	err    error
}

func (w *failAfterWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.failOn > 0 && w.writes >= w.failOn {
		if w.err == nil {
			w.err = io.ErrClosedPipe
		}
		return 0, w.err
	}
	return len(p), nil
}
