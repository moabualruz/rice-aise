package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mkh/rice-aise/internal/domain"
	"github.com/mkh/rice-aise/internal/platform"
	"github.com/mkh/rice-aise/internal/storage"
)

const vanillaProfileName = "vanilla"

var nowFunc = time.Now

type ProfileService struct {
	store    storage.ProfileStore
	resolver platform.PathResolver
	registry []domain.Tool
}

func NewProfileService(
	store storage.ProfileStore,
	resolver platform.PathResolver,
	registry []domain.Tool,
) *ProfileService {
	return &ProfileService{store: store, resolver: resolver, registry: registry}
}

func (s *ProfileService) DetectInstalled() []domain.Tool {
	tools := []domain.Tool{}
	for _, tool := range s.registry {
		if len(tool.ConfigDirs) == 0 {
			continue
		}
		path := s.resolver.ResolvePath(tool.ConfigDirs[0].SourcePath)
		if s.resolver.Exists(path) {
			tools = append(tools, tool)
		}
	}
	return tools
}

func (s *ProfileService) Init() (string, error) {
	if err := s.ensureUninitialized(); err != nil {
		return "", err
	}
	tools := s.DetectInstalled()
	current := nowFunc().Format("2006-01-02") + "-current"
	if _, err := s.store.Create(current, toolMap(tools)); err != nil {
		return "", err
	}
	if err := s.captureCurrentProfile(current, tools); err != nil {
		return "", err
	}
	if err := s.createVanillaProfile(current, tools); err != nil {
		return "", err
	}
	return current, s.store.SetActive(activeMap(tools, current))
}

func (s *ProfileService) Save(name string) error {
	active, err := s.store.GetActive()
	if err != nil {
		return err
	}
	tools, err := s.toolsFromActive(active)
	if err != nil {
		return err
	}
	if _, err := s.store.Create(name, toolMap(tools)); err != nil {
		return err
	}
	return s.copyLiveTools(s.store.ProfileDir(name), tools)
}

func (s *ProfileService) Use(profileName string, onlyTools []string) error {
	profile, err := s.store.Get(profileName)
	if err != nil {
		return err
	}
	active, err := s.store.GetActive()
	if err != nil {
		return err
	}
	tools, err := s.selectTools(profile, onlyTools)
	if err != nil {
		return err
	}
	if err := s.linkProfileTools(profileName, tools); err != nil {
		return err
	}
	for _, tool := range tools {
		active[tool.ID] = profileName
	}
	return s.store.SetActive(active)
}

func (s *ProfileService) Create(name string) error {
	tools := s.DetectInstalled()
	active, err := s.store.GetActive()
	if err != nil {
		return err
	}
	if _, err := s.store.Create(name, toolMap(tools)); err != nil {
		return err
	}
	return s.seedProfile(name, tools, active)
}

func (s *ProfileService) Which(toolID string) (string, error) {
	active, err := s.store.GetActive()
	if err != nil {
		return "", err
	}
	name, ok := active[toolID]
	if !ok {
		return "", fmt.Errorf("tool %s has no active profile", toolID)
	}
	return name, nil
}

func (s *ProfileService) ensureUninitialized() error {
	profiles, err := s.store.List()
	if err != nil {
		return err
	}
	if len(profiles) != 0 {
		return fmt.Errorf("already initialized")
	}
	return nil
}

func (s *ProfileService) captureCurrentProfile(name string, tools []domain.Tool) error {
	for _, tool := range tools {
		if err := s.captureCurrentTool(name, tool); err != nil {
			return err
		}
	}
	return nil
}

func (s *ProfileService) captureCurrentTool(name string, tool domain.Tool) error {
	if err := s.ensureToolNotLinked(tool); err != nil {
		return err
	}
	if err := s.moveCurrentDirs(name, tool); err != nil {
		return err
	}
	if !tool.DotfilesPattern {
		return nil
	}
	return s.moveCurrentDotfiles(name, tool)
}

func (s *ProfileService) ensureToolNotLinked(tool domain.Tool) error {
	for i := range tool.ConfigDirs {
		path := s.resolver.ResolveToolDir(tool, i)
		if s.resolver.IsSymlink(path) {
			return fmt.Errorf("%s already linked", path)
		}
	}
	return nil
}

func (s *ProfileService) moveCurrentDirs(name string, tool domain.Tool) error {
	for i, dir := range tool.ConfigDirs {
		livePath := s.resolver.ResolveToolDir(tool, i)
		target := profileConfigPath(s.store.ProfileDir(name), dir)
		if err := moveOrSeedDir(livePath, target); err != nil {
			return err
		}
		if err := swapSymlink(target, livePath); err != nil {
			return err
		}
	}
	return nil
}

func (s *ProfileService) moveCurrentDotfiles(name string, tool domain.Tool) error {
	for _, dotfile := range tool.Dotfiles {
		livePath := homePath(s.resolver.HomeDir(), dotfile)
		target := profileDotfilePath(s.store.ProfileDir(name), tool, dotfile)
		if err := moveOrSeedFile(livePath, target); err != nil {
			return err
		}
		if err := swapSymlink(target, livePath); err != nil {
			return err
		}
	}
	return nil
}

func (s *ProfileService) createVanillaProfile(current string, tools []domain.Tool) error {
	if _, err := s.store.Create(vanillaProfileName, toolMap(tools)); err != nil {
		return err
	}
	return s.seedProfile(vanillaProfileName, tools, activeMap(tools, current))
}

func (s *ProfileService) seedProfile(
	name string,
	tools []domain.Tool,
	active map[string]string,
) error {
	profileDir := s.store.ProfileDir(name)
	for _, tool := range tools {
		if err := ensureProfileLayout(profileDir, tool); err != nil {
			return err
		}
		srcName, ok := active[tool.ID]
		if !ok {
			continue
		}
		if err := copyCredentials(s.store.ProfileDir(srcName), profileDir, tool); err != nil {
			return err
		}
	}
	return nil
}

func (s *ProfileService) toolsFromActive(active map[string]string) ([]domain.Tool, error) {
	tools := []domain.Tool{}
	for id := range active {
		tool, err := s.requireTool(id)
		if err != nil {
			return nil, err
		}
		tools = append(tools, tool)
	}
	return tools, nil
}

func (s *ProfileService) copyLiveTools(profileDir string, tools []domain.Tool) error {
	for _, tool := range tools {
		if err := s.copyLiveTool(profileDir, tool); err != nil {
			return err
		}
	}
	return nil
}

func (s *ProfileService) copyLiveTool(profileDir string, tool domain.Tool) error {
	if err := ensureProfileLayout(profileDir, tool); err != nil {
		return err
	}
	if err := s.copyLiveDirs(profileDir, tool); err != nil {
		return err
	}
	if !tool.DotfilesPattern {
		return nil
	}
	return s.copyLiveDotfiles(profileDir, tool)
}

func (s *ProfileService) copyLiveDirs(profileDir string, tool domain.Tool) error {
	for i, dir := range tool.ConfigDirs {
		src, err := resolveCopySource(s.resolver.ResolveToolDir(tool, i))
		if err != nil {
			return err
		}
		if err := copyEntry(src, profileConfigPath(profileDir, dir)); err != nil {
			return err
		}
	}
	return nil
}

func (s *ProfileService) copyLiveDotfiles(profileDir string, tool domain.Tool) error {
	for _, dotfile := range tool.Dotfiles {
		src, err := resolveCopySource(homePath(s.resolver.HomeDir(), dotfile))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if err := copyEntry(src, profileDotfilePath(profileDir, tool, dotfile)); err != nil {
			return err
		}
	}
	return nil
}

func (s *ProfileService) selectTools(
	profile *domain.Profile,
	onlyTools []string,
) ([]domain.Tool, error) {
	ids := onlyTools
	if len(ids) == 0 {
		ids = enabledToolIDs(profile)
	}
	return s.requireTools(profile, ids)
}

func (s *ProfileService) requireTools(
	profile *domain.Profile,
	ids []string,
) ([]domain.Tool, error) {
	tools := make([]domain.Tool, 0, len(ids))
	for _, id := range ids {
		if !profile.Tools[id] {
			return nil, fmt.Errorf("profile %s does not include %s", profile.Name, id)
		}
		tool, err := s.requireTool(id)
		if err != nil {
			return nil, err
		}
		tools = append(tools, tool)
	}
	return tools, nil
}

func (s *ProfileService) requireTool(id string) (domain.Tool, error) {
	tool := domain.FindTool(s.registry, id)
	if tool == nil {
		return domain.Tool{}, fmt.Errorf("unknown tool %s", id)
	}
	return *tool, nil
}

func (s *ProfileService) linkProfileTools(name string, tools []domain.Tool) error {
	for _, tool := range tools {
		if err := s.linkProfileTool(name, tool); err != nil {
			return err
		}
	}
	return nil
}

func (s *ProfileService) linkProfileTool(name string, tool domain.Tool) error {
	if err := s.linkProfileDirs(name, tool); err != nil {
		return err
	}
	if !tool.DotfilesPattern {
		return nil
	}
	return s.linkProfileDotfiles(name, tool)
}

func (s *ProfileService) linkProfileDirs(name string, tool domain.Tool) error {
	for i, dir := range tool.ConfigDirs {
		target := profileConfigPath(s.store.ProfileDir(name), dir)
		linkPath := s.resolver.ResolveToolDir(tool, i)
		if err := swapSymlink(target, linkPath); err != nil {
			return err
		}
	}
	return nil
}

func (s *ProfileService) linkProfileDotfiles(name string, tool domain.Tool) error {
	for _, dotfile := range tool.Dotfiles {
		target := profileDotfilePath(s.store.ProfileDir(name), tool, dotfile)
		linkPath := homePath(s.resolver.HomeDir(), dotfile)
		if err := swapSymlink(target, linkPath); err != nil {
			return err
		}
	}
	return nil
}

var credentialPathFunc = credentialPath

func copyCredentials(srcProfileDir, dstProfileDir string, tool domain.Tool) error {
	for _, credential := range tool.CredentialFiles {
		src, err := credentialPathFunc(srcProfileDir, tool, credential)
		if err != nil {
			return err
		}
		dst, err := credentialPathFunc(dstProfileDir, tool, credential)
		if err != nil {
			return err
		}
		if err := copyCredentialFile(src, dst); err != nil {
			return err
		}
	}
	return nil
}

func toolMap(tools []domain.Tool) map[string]bool {
	mapping := map[string]bool{}
	for _, tool := range tools {
		mapping[tool.ID] = true
	}
	return mapping
}

func activeMap(tools []domain.Tool, profileName string) map[string]string {
	mapping := map[string]string{}
	for _, tool := range tools {
		mapping[tool.ID] = profileName
	}
	return mapping
}

func enabledToolIDs(profile *domain.Profile) []string {
	ids := []string{}
	for id, enabled := range profile.Tools {
		if enabled {
			ids = append(ids, id)
		}
	}
	return ids
}

func profileConfigPath(profileDir string, dir domain.DirMapping) string {
	return filepath.Join(profileDir, dir.ProfileSubdir)
}

func profileDotfilePath(profileDir string, tool domain.Tool, dotfile string) string {
	return filepath.Join(profileDir, tool.ID, filepath.Clean(dotfile))
}

func homePath(homeDir, relative string) string {
	return filepath.Join(homeDir, filepath.Clean(relative))
}

func ensureProfileLayout(profileDir string, tool domain.Tool) error {
	for _, dir := range tool.ConfigDirs {
		if err := os.MkdirAll(profileConfigPath(profileDir, dir), 0o755); err != nil {
			return err
		}
	}
	if !tool.DotfilesPattern {
		return nil
	}
	return ensureDotfileLayout(profileDir, tool)
}

func ensureDotfileLayout(profileDir string, tool domain.Tool) error {
	for _, dotfile := range tool.Dotfiles {
		if err := writeEmptyFile(profileDotfilePath(profileDir, tool, dotfile)); err != nil {
			return err
		}
	}
	return nil
}

func moveOrSeedDir(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return os.MkdirAll(dst, 0o755)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.Rename(src, dst)
}

func moveOrSeedFile(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return writeEmptyFile(dst)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.Rename(src, dst)
}

func swapSymlink(target, linkPath string) error {
	tmpPath := linkPath + ".tmp"
	_ = os.Remove(tmpPath)
	if err := os.MkdirAll(filepath.Dir(linkPath), 0o755); err != nil {
		return err
	}
	if err := os.Symlink(target, tmpPath); err != nil {
		return err
	}
	return os.Rename(tmpPath, linkPath)
}

func credentialPath(profileDir string, tool domain.Tool, credential string) (string, error) {
	if !strings.HasPrefix(credential, "~") && !filepath.IsAbs(credential) {
		return filepath.Join(profileDir, tool.ConfigDirs[0].ProfileSubdir, credential), nil
	}
	if tool.DotfilesPattern && strings.HasPrefix(credential, "~/") {
		return filepath.Join(profileDir, tool.ID, credential[2:]), nil
	}
	for _, dir := range tool.ConfigDirs {
		if path, ok := dirCredentialPath(profileDir, dir, credential); ok {
			return path, nil
		}
	}
	return "", fmt.Errorf("credential %q outside managed paths", credential)
}

func dirCredentialPath(
	profileDir string,
	dir domain.DirMapping,
	credential string,
) (string, bool) {
	source := filepath.Clean(dir.SourcePath)
	cleanCredential := filepath.Clean(credential)
	if cleanCredential != source && !strings.HasPrefix(cleanCredential, source+string(os.PathSeparator)) {
		return "", false
	}
	relative, _ := filepath.Rel(source, cleanCredential)
	return filepath.Join(profileDir, dir.ProfileSubdir, relative), true
}

func copyCredentialFile(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}
	return copyFile(src, dst)
}

func resolveCopySource(path string) (string, error) {
	if _, err := os.Lstat(path); err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved, nil
	}
	return path, nil
}

func copyEntry(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyDir(src, dst string) error {
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return copyWalkEntry(src, dst, path, info)
	})
}

func copyWalkEntry(src, dst, path string, info os.FileInfo) error {
	relative, _ := filepath.Rel(src, path)
	target := filepath.Join(dst, relative)
	if info.IsDir() {
		return os.MkdirAll(target, info.Mode())
	}
	return copyFile(path, target)
}

func copyFile(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	reader, err := os.Open(src)
	if err != nil {
		return err
	}
	defer reader.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	writer, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
	if err != nil {
		return err
	}
	defer writer.Close()
	_, err = io.Copy(writer, reader)
	return err
}

func writeEmptyFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, nil, 0o644)
}
