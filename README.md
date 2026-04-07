<div align="center">

<img src=".branding/banner.png" alt="riceAise" width="500">

# raise

[![License: CC BY-NC-SA 4.0](https://img.shields.io/badge/License-CC%20BY--NC--SA%204.0-lightgrey.svg)](https://creativecommons.org/licenses/by-nc-sa/4.0/)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![CI](https://github.com/moabualruz/rice-aise/actions/workflows/ci.yml/badge.svg)](https://github.com/moabualruz/rice-aise/actions)

**Switch between AI agent configurations in one command.**

*Full setup with plugins and hooks, vanilla clean slate, or any custom profile — instantly, for 17 AI coding tools at once.*

[Quickstart](#quickstart) · [Commands](#commands) · [Supported Tools](#supported-tools) · [How It Works](#how-it-works) · [Contributing](CONTRIBUTING.md)

</div>

---

## The Problem

You have Claude Code with 20 plugins, Codex with custom agents, Gemini with MCP servers — and sometimes you just want a clean slate. Or you want to test a new setup without losing your current one. Or you have a "work" config and a "personal" config.

Today, switching means manually moving dozens of directories and hoping you remember where everything was.

**raise** solves this with profile-based config switching via symlinks — atomic, instant, and credential-preserving.

```
~/.claude/ ──symlink──> ~/.raise/profiles/full-setup/claude/
~/.codex/  ──symlink──> ~/.raise/profiles/full-setup/codex/
~/.gemini/ ──symlink──> ~/.raise/profiles/full-setup/gemini/
...

$ raise use vanilla    # instant switch, credentials preserved

~/.claude/ ──symlink──> ~/.raise/profiles/vanilla/claude/
~/.codex/  ──symlink──> ~/.raise/profiles/vanilla/codex/
~/.gemini/ ──symlink──> ~/.raise/profiles/vanilla/gemini/
```

## Quickstart

### Install

```bash
# From source
git clone https://github.com/moabualruz/rice-aise.git
cd rice-aise
go build -o raise ./cmd/raise
sudo mv raise /usr/local/bin/

# Or with go install
go install github.com/mkh/rice-aise/cmd/raise@latest
```

### First Run

```bash
# Detect installed tools, save current config, create vanilla profile
raise init

# See what was detected
raise status

# Switch to vanilla (clean config, your credentials preserved)
raise use vanilla

# Switch back to your full setup
raise use 2026-04-07-current
```

## Commands

| Command | Description |
|---------|-------------|
| `raise init` | First-run setup — detect tools, save current config, create vanilla profile |
| `raise use <name>` | Switch all tools to a named profile |
| `raise use <name> --only tool1,tool2` | Switch only specified tools |
| `raise list` | List all profiles (`*` marks active) |
| `raise status` | Show active profile per tool in table format |
| `raise create <name>` | Create new profile with migrated credentials |
| `raise save <name>` | Snapshot current live config into a profile |
| `raise rename <old> <new>` | Rename a profile |
| `raise delete <name>` | Delete a profile (prompts for confirmation) |
| `raise delete <name> --force` | Delete without confirmation |
| `raise tools` | List all 17 supported tools and their config paths |
| `raise which <tool>` | Show which profile a specific tool uses |
| `raise version` | Print version |

### Examples

```bash
# Switch only Claude and Codex to vanilla, keep everything else
raise use vanilla --only claude,codex

# Create a new "experiment" profile from current credentials
raise create experiment

# Check what Claude is currently using
raise which claude

# Snapshot your current tweaked config before trying something new
raise save before-experiment
```

## Supported Tools

| ID | Tool | Config Directories |
|----|------|--------------------|
| `claude` | Claude Code | `~/.claude/` |
| `codex` | Codex CLI (OpenAI) | `~/.codex/` |
| `gemini` | Gemini CLI (Google) | `~/.gemini/` |
| `opencode` | OpenCode | `~/.config/opencode/` `~/.local/share/opencode/` `~/.cache/opencode/` `~/.local/state/opencode/` |
| `pi` | Pi Coding Agent | `~/.pi/` |
| `copilot` | GitHub Copilot CLI | `~/.config/github-copilot/` `~/.copilot/` |
| `omc` | oh-my-claudecode | `~/.omc/` |
| `omx` | oh-my-codex | `~/.omx/` |
| `omg` | oh-my-gemini | `~/.omg/` |
| `aider` | Aider | `~/.aider/` + home dotfiles |
| `cline` | Cline CLI | `~/.cline/` |
| `continue` | Continue CLI | `~/.continue/` |
| `amp` | Amp (Sourcegraph) | `~/.config/amp/` |
| `goose` | Goose (Block) | `~/.config/block/goose/` `~/.local/share/goose/` `~/.local/state/goose/` |
| `amazonq` | Amazon Q Developer | `~/.aws/amazonq/` |
| `kiro` | Kiro CLI (AWS) | `~/.kiro/` |
| `kilo` | Kilo Code CLI | `~/.config/kilo/` |

## How It Works

### Init

`raise init` performs a one-time setup:

1. **Detects** which AI tools are installed by checking for their config directories
2. **Moves** all config directories into `~/.raise/profiles/<dated-name>/`
3. **Creates symlinks** from the original paths back to the profile
4. **Creates a vanilla profile** containing only your credential files (OAuth tokens, API keys) — no plugins, hooks, or extensions
5. **Sets the current profile** as active

### Switching

`raise use <name>` atomically re-points symlinks using a safe temp-symlink + rename pattern:

```
os.Symlink(target, path + ".tmp")   # create new symlink
os.Rename(path + ".tmp", path)       # atomic swap
```

### Credential Preservation

When creating profiles (`raise init`, `raise create`), only authentication files are migrated:

| Tool | Credential Files |
|------|-----------------|
| Claude | `.credentials.json` |
| Codex | `auth.json` |
| Gemini | `oauth_creds.json`, `google_accounts.json`, `mcp-oauth-tokens.json` |
| OpenCode | `auth.json` (in data dir) |
| Pi | `agent/auth.json` |
| Copilot | `apps.json` (in auth dir) |
| Cline | `data/secrets.json` |
| Continue | `config.yaml` (keys inline) |
| Aider | `.aider.env` |

You never need to re-authenticate after switching profiles.

### Storage Layout

```
~/.raise/
  config.json                      # active profile mapping per tool
  profiles/
    2026-04-07-current/            # your original full config
      claude/                      # full ~/.claude/ contents
      codex/                       # full ~/.codex/ contents
      gemini/                      # full ~/.gemini/ contents
      opencode-config/             # ~/.config/opencode/
      opencode-data/               # ~/.local/share/opencode/
      opencode-cache/              # ~/.cache/opencode/
      opencode-state/              # ~/.local/state/opencode/
      copilot-auth/                # ~/.config/github-copilot/
      copilot-config/              # ~/.copilot/
      ...
    vanilla/                       # clean config (credentials only)
      claude/
      codex/
      ...
```

## Special Cases

| Tool | Handling |
|------|----------|
| **Aider** | Uses home-directory dotfiles (`.aider.conf.yml`, `.aider.env`, etc.) — each gets its own symlink rather than a directory swap |
| **OpenCode** | Manages 4 XDG directories atomically — config, data, cache, and state |
| **Copilot** | Manages 2 directories — auth (`~/.config/github-copilot/`) and config (`~/.copilot/`) |
| **Goose** | Config dirs are managed; credentials use OS keyring (not file-swappable) |
| **Amazon Q / Kiro** | Only manages tool-specific dirs (`~/.aws/amazonq/`, `~/.kiro/`) — shared AWS credentials (`~/.aws/credentials`) are never touched |

## Architecture

```
cmd/raise/              Entry point (3 lines)
internal/
  domain/               Tool, Profile, DirMapping entities
                        ToolRegistry with all 17 tool definitions
  platform/             PathResolver — tilde expansion, env overrides
  storage/              ProfileStore interface + FileProfileStore (JSON)
  service/              ProfileService — init, save, use, create, switch
                        Symlink mechanics, credential migration
  cli/                  Cobra commands — thin wrappers over service
tests/
  integration/          Real filesystem lifecycle tests
  e2e/                  Binary subprocess tests
```

**Design principles:** DDD (domain-driven), SOLID interfaces, DRY registry, TDD throughout.

**Test coverage:** domain 100% | platform 100% | service 100% | storage 100% | cli 98.9%

## Development

```bash
# All unit tests
go test ./...

# Integration tests (real filesystem operations)
go test ./tests/integration/... -tags integration -v

# E2E tests (builds and runs the actual binary)
go test ./tests/e2e/... -tags e2e -v

# Full verification
go vet ./... && go test ./... && \
  go test ./tests/integration/... -tags integration && \
  go test ./tests/e2e/... -tags e2e

# Build
go build -o raise ./cmd/raise
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines. Adding a new tool is straightforward — one function in `registry.go` and matching tests.

## License

[CC BY-NC-SA 4.0](LICENSE.md) - Copyright (c) 2026 RiceCoder Contributors

---

<div align="center">
<sub>Part of the <b>rice</b> ecosystem</sub>
</div>
