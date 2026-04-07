package storage

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/mkh/rice-aise/internal/domain"
)

type ProfileStore interface {
	Create(name string, tools map[string]bool) (*domain.Profile, error)
	Get(name string) (*domain.Profile, error)
	List() ([]domain.Profile, error)
	Delete(name string) error
	Rename(oldName, newName string) error
	SetActive(mapping map[string]string) error
	GetActive() (map[string]string, error)
	ProfileDir(name string) string
	RaiseDir() string
}

type FileProfileStore struct {
	raiseDir string
}

type metadataFile struct {
	Name      string          `json:"name"`
	CreatedAt string          `json:"createdAt"`
	Tools     map[string]bool `json:"tools"`
}

type configFile struct {
	Active map[string]string `json:"active"`
}

func NewFileProfileStore(raiseDir string) *FileProfileStore {
	return &FileProfileStore{raiseDir: raiseDir}
}

func (s *FileProfileStore) Create(name string, tools map[string]bool) (*domain.Profile, error) {
	profile := domain.Profile{
		Name:      name,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		Tools:     maps.Clone(tools),
	}
	if err := s.writeMetadata(profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func (s *FileProfileStore) Get(name string) (*domain.Profile, error) {
	meta := metadataFile{}
	if err := s.readJSON(s.metadataPath(name), &meta); err != nil {
		return nil, err
	}
	return meta.profile()
}

func (s *FileProfileStore) List() ([]domain.Profile, error) {
	entries, err := os.ReadDir(s.profilesDir())
	if os.IsNotExist(err) {
		return []domain.Profile{}, nil
	}
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	slices.Sort(names)

	list := make([]domain.Profile, 0, len(names))
	for _, name := range names {
		profile, err := s.Get(name)
		if err != nil {
			return nil, err
		}
		list = append(list, *profile)
	}
	return list, nil
}

func (s *FileProfileStore) Delete(name string) error {
	return os.RemoveAll(s.ProfileDir(name))
}

func (s *FileProfileStore) Rename(oldName, newName string) error {
	if err := os.Rename(s.ProfileDir(oldName), s.ProfileDir(newName)); err != nil {
		return err
	}
	profile, err := s.Get(newName)
	if err != nil {
		return err
	}
	profile.Name = newName
	return s.writeMetadata(*profile)
}

func (s *FileProfileStore) SetActive(mapping map[string]string) error {
	cfg := configFile{Active: maps.Clone(mapping)}
	return s.writeJSON(s.configPath(), cfg)
}

func (s *FileProfileStore) GetActive() (map[string]string, error) {
	cfg := configFile{}
	err := s.readJSON(s.configPath(), &cfg)
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	return maps.Clone(cfg.Active), nil
}

func (s *FileProfileStore) ProfileDir(name string) string {
	return filepath.Join(s.profilesDir(), name)
}

func (s *FileProfileStore) RaiseDir() string {
	return s.raiseDir
}

func (s *FileProfileStore) profilesDir() string {
	return filepath.Join(s.raiseDir, "profiles")
}

func (s *FileProfileStore) metadataPath(name string) string {
	return filepath.Join(s.ProfileDir(name), "metadata.json")
}

func (s *FileProfileStore) configPath() string {
	return filepath.Join(s.raiseDir, "config.json")
}

func (s *FileProfileStore) writeMetadata(profile domain.Profile) error {
	meta := metadataFile{
		Name:      profile.Name,
		CreatedAt: profile.CreatedAt.UTC().Format(time.RFC3339),
		Tools:     maps.Clone(profile.Tools),
	}
	return s.writeJSON(s.metadataPath(profile.Name), meta)
}

func (m metadataFile) profile() (*domain.Profile, error) {
	createdAt, err := time.Parse(time.RFC3339, m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse createdAt: %w", err)
	}
	return &domain.Profile{
		Name:      m.Name,
		CreatedAt: createdAt,
		Tools:     maps.Clone(m.Tools),
	}, nil
}

func (s *FileProfileStore) readJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func (s *FileProfileStore) writeJSON(path string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
