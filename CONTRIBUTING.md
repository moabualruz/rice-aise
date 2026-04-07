# Contributing to raise

Thank you for your interest in contributing! This document provides quick guidelines.

## Getting Started

```bash
# Clone and setup
git clone https://github.com/moabualruz/rice-aise.git
cd rice-aise

# Build
go build -o raise ./cmd/raise

# Run unit tests
go test ./...

# Run integration tests
go test ./tests/integration/... -tags integration -v

# Run E2E tests
go test ./tests/e2e/... -tags e2e -v

# Vet
go vet ./...
```

## Ways to Contribute

### Report Bugs

1. Check [existing issues](https://github.com/moabualruz/rice-aise/issues)
2. Create new issue with:
   - `raise version` output
   - Steps to reproduce
   - Expected vs actual behavior

### Suggest Features

1. Open a [Discussion](https://github.com/moabualruz/rice-aise/discussions)
2. Describe use case and benefits
3. Consider implementation approach

### Submit Code

1. Fork the repository
2. Create feature branch: `git checkout -b feat/my-feature`
3. Write tests first (TDD)
4. Run checks: `go test ./... && go vet ./...`
5. Submit pull request

## Code Style

- Go 1.22+ with idiomatic patterns
- Use [Conventional Commits](https://www.conventionalcommits.org/)
- Keep functions under 30 lines
- Files under 800 lines
- 100% test coverage on new code

## Pull Request Checklist

- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes
- [ ] Integration tests pass (`-tags integration`)
- [ ] E2E tests pass (`-tags e2e`)
- [ ] Tests added for new code
- [ ] Documentation updated if needed
- [ ] Commit messages follow conventions

## Adding a New Tool

1. Add tool definition in `internal/domain/registry.go` — implement a function returning `Tool{}` with correct config dirs, credential files, and env override
2. Register it in `NewToolRegistry()` 
3. Add tests in `internal/domain/registry_test.go` verifying all fields
4. Update `docs/tool-registry-reference.md` with the full config/credential documentation
5. Update `README.md` supported tools table

## Project Structure

```
cmd/raise/          — CLI entry point
internal/
  domain/           — Tool, Profile, DirMapping entities + registry
  platform/         — PathResolver, tilde expansion, env overrides
  storage/          — ProfileStore interface + FileProfileStore (JSON)
  service/          — ProfileService (init, save, use, create, switch)
  cli/              — Cobra commands
tests/
  integration/      — Real filesystem lifecycle tests
  e2e/              — Binary subprocess tests
```

## Code of Conduct

Be respectful and inclusive. We welcome contributors of all backgrounds and experience levels.

## License

By contributing, you agree your contributions will be licensed under [CC BY-NC-SA 4.0](https://creativecommons.org/licenses/by-nc-sa/4.0/).

## Questions?

- Open a [Discussion](https://github.com/moabualruz/rice-aise/discussions)
