# raise — Tool Registry Reference

Verified against official documentation (Context7) and local filesystem as of 2026-04-07.

## Legend

- **Config Dir**: directories managed by `raise` (swapped via symlinks)
- **Credential Files**: files that MUST be migrated when creating vanilla profiles (relative to config dir unless absolute)
- **Env Override**: environment variable to override config dir location
- **Platform Notes**: OS-specific differences

---

## 1. Claude Code

| Field | Value |
|-------|-------|
| ID | `claude` |
| Config Dirs | `~/.claude/` |
| Env Override | `CLAUDE_CONFIG_DIR` |
| Credential Files | `.credentials.json` |
| Credential Notes | macOS: stored in encrypted Keychain. Linux/Windows: stored in `.credentials.json`. Supports `apiKeyHelper` script for dynamic keys. API keys via `ANTHROPIC_API_KEY` or `ANTHROPIC_AUTH_TOKEN` env vars. |
| Settings Files | `settings.json`, `settings.local.json`, `keybindings.json` |
| State Files | `history.jsonl`, `sessions/`, `metrics/costs.jsonl`, `cache/`, `projects/`, `paste-cache/`, `image-cache/`, `shell-snapshots/`, `session-env/`, `telemetry/`, `mcp-health-cache.json` |
| Extension Files | `hooks/`, `commands/`, `agents/`, `skills/`, `plugins/`, `plugin.json`, `CLAUDE.md`, `mcp-configs/` |
| Platform | All platforms: `~/.claude/`. Windows: `%APPDATA%\Claude\` (alternate). |

---

## 2. Codex CLI (OpenAI)

| Field | Value |
|-------|-------|
| ID | `codex` |
| Config Dirs | `~/.codex/` |
| Env Override | `CODEX_HOME` |
| Credential Files | `auth.json` |
| Credential Notes | `auth.json` holds OAuth tokens (id_token, access_token, refresh_token) from ChatGPT login. Also supports `.env` file for API keys. Alternative: system keyring when `auth_credentials_store_mode = "keyring"`. |
| Settings Files | `config.toml`, `AGENTS.md`, `AGENTS.override.md`, `hooks.json` |
| State Files | `history.jsonl`, `log/`, `sessions/` (date-organized rollouts), SQLite state DB |
| Extension Files | `skills/`, `plugins/`, `proxy/` |
| Platform | All platforms: `~/.codex/`. No XDG support. |

---

## 3. Gemini CLI (Google)

| Field | Value |
|-------|-------|
| ID | `gemini` |
| Config Dirs | `~/.gemini/` |
| Env Override | `GEMINI_CLI_HOME` |
| Credential Files | `oauth_creds.json`, `google_accounts.json`, `mcp-oauth-tokens.json` |
| Credential Notes | `oauth_creds.json` holds Google OAuth credentials. `google_accounts.json` holds account info. `mcp-oauth-tokens.json` holds MCP server OAuth tokens (auto-managed). API key via `GEMINI_API_KEY` env var or `.env` file. |
| Settings Files | `settings.json`, `trustedFolders.json`, `GEMINI.md` |
| State Files | `tmp/<project_hash>/shell_history`, `tmp/<project_hash>/chats/`, `history/`, `projects.json`, `state.json`, `installation_id` |
| Extension Files | `extensions/` |
| Platform | All platforms: `~/.gemini/`. System defaults: Linux `/etc/gemini-cli/system-defaults.json`, macOS `/Library/Application Support/GeminiCli/system-defaults.json`, Windows `C:\ProgramData\gemini-cli\settings.json`. |

---

## 4. OpenCode

| Field | Value |
|-------|-------|
| ID | `opencode` |
| Config Dirs | `~/.config/opencode/`, `~/.local/share/opencode/`, `~/.cache/opencode/`, `~/.local/state/opencode/` |
| Env Override | `OPENCODE_CONFIG_DIR` (config), `OPENCODE_CONFIG` (specific file) |
| Credential Files | `~/.local/share/opencode/auth.json` |
| Credential Notes | `auth.json` in the **data** dir (not config dir) holds API keys and OAuth tokens. Created via `opencode auth login`. Also loads from env vars and `.env` files. |
| Settings Files | `~/.config/opencode/opencode.json` (or `.jsonc`), plus variant configs (`oh-my-opencode.jsonc`, `AGENTS.md`) |
| State Files | `~/.local/share/opencode/storage/` (sessions, messages, parts, projects), `~/.local/share/opencode/log/`, `~/.local/share/opencode/opencode.db`, `~/.local/share/opencode/snapshot/`, `~/.local/share/opencode/tool-output/`, `~/.local/share/opencode/delegations/` |
| Cache Files | `~/.cache/opencode/version`, `~/.cache/opencode/bin/`, `~/.cache/opencode/models.json`, `~/.cache/opencode/node_modules/`, `~/.cache/opencode/packages/` |
| State Dir Files | `~/.local/state/opencode/prompt-history.jsonl`, `~/.local/state/opencode/model.json`, `~/.local/state/opencode/locks/` |
| Extension Files | `~/.local/share/opencode/plugins/` |
| Platform | Uses XDG base directories. macOS/Linux: as above. Windows: `%USERPROFILE%\.local\share\opencode\`. |

---

## 5. Pi Coding Agent

| Field | Value |
|-------|-------|
| ID | `pi` |
| Config Dirs | `~/.pi/` |
| Env Override | `PI_CONFIG_DIR` |
| Credential Files | `agent/auth.json` |
| Credential Notes | `agent/auth.json` holds API keys/tokens. Can also set keys in `agent/settings.json` or via env vars. |
| Settings Files | `agent/settings.json` |
| State Files | Session data (location controlled by `sessionDir` in settings.json) |
| Extension Files | `agent/extensions/` |
| Platform | All platforms: `~/.pi/`. XDG support via `$XDG_CONFIG_HOME/pi/` (after fix #2870). |

---

## 6. GitHub Copilot CLI

| Field | Value |
|-------|-------|
| ID | `copilot` |
| Config Dirs | `~/.config/github-copilot/`, `~/.copilot/` |
| Env Override | `COPILOT_HOME` (for `~/.copilot/`) |
| Credential Files | `~/.config/github-copilot/apps.json` |
| Credential Notes | `apps.json` holds GitHub OAuth token (`ghu_…`) keyed by App ID. Auth can also be done via `GH_TOKEN`, `GITHUB_TOKEN`, or `COPILOT_GITHUB_TOKEN` env vars. No credentials stored in `~/.copilot/`. |
| Settings Files | `~/.copilot/config.json`, `~/.copilot/mcp-config.json`, `~/.copilot/AGENTS.md`, `~/.copilot/PROJECTS.md` |
| State Files | `~/.copilot/history-session-state/`, `~/.copilot/ide-state/`, `~/.copilot/permissions.json`, `~/.copilot/logs/` |
| Extension Files | `~/.copilot/agents/`, `~/.copilot/skills/`, `~/.copilot/hooks/`, `~/.copilot/installed-plugins/`, `~/.copilot/knowledge/`, `~/.copilot/templates/`, `~/.copilot/rules/`, `~/.copilot/projects/` |
| Platform | All platforms: `~/.copilot/` and `~/.config/github-copilot/`. Same on macOS/Linux/Windows. |

---

## 7. oh-my-claudecode (OMC)

| Field | Value |
|-------|-------|
| ID | `omc` |
| Config Dirs | `~/.omc/` |
| Env Override | none |
| Credential Files | none |
| Settings Files | none (settings managed within `~/.claude/`) |
| State Files | `state/` (ralph-state.json, ultrawork-state.json, mission-state.json, subagent-tracking.json, agent-replay-*.jsonl, last-tool-error.json, idle-notif-cooldown.json, sessions/), `sessions/` |
| Platform | All platforms: `~/.omc/`. |

---

## 8. oh-my-codex (OMX)

| Field | Value |
|-------|-------|
| ID | `omx` |
| Config Dirs | `~/.omx/` |
| Env Override | none |
| Credential Files | none |
| Settings Files | `hud-config.json`, `setup-scope.json` |
| State Files | `state/`, `logs/`, `metrics.json`, `backups/`, `plans/` |
| Platform | All platforms: `~/.omx/`. |

---

## 9. oh-my-gemini (OMG)

| Field | Value |
|-------|-------|
| ID | `omg` |
| Config Dirs | `~/.omg/` |
| Env Override | none |
| Credential Files | none |
| Settings Files | `setup-scope.json` |
| State Files | `state/` |
| Platform | All platforms: `~/.omg/`. |

---

## 10. Aider

| Field | Value |
|-------|-------|
| ID | `aider` |
| Config Dirs | `~/.aider/` (cache), plus dotfiles in `~/` |
| Env Override | `--env-file <path>` flag |
| Credential Files | `~/.aider.env` |
| Credential Notes | `~/.aider.env` holds API keys (OPENAI_API_KEY, ANTHROPIC_API_KEY, etc.). Loaded as env vars. Project-level `.env` also loaded. Keys can also be set in `~/.aider.conf.yml` (openai-api-key, anthropic-api-key fields). |
| Settings Files | `~/.aider.conf.yml`, `~/.aider.model.settings.yml`, `~/.aider.model.metadata.json` |
| State Files | `~/.aider/` (chat history, session data), `.aider.chat.history.md` (per-project), `.aider.input.history` (per-project), `.aider.tags.cache.v3/` (per-project) |
| Dotfiles Pattern | Aider uses **home-directory dotfiles** rather than a config directory. Files to manage: `~/.aider.conf.yml`, `~/.aider.env`, `~/.aider.model.settings.yml`, `~/.aider.model.metadata.json`, `~/.aider/` |
| Platform | All platforms: `~/` dotfiles + `~/.aider/`. |

---

## 11. Cline CLI

| Field | Value |
|-------|-------|
| ID | `cline` |
| Config Dirs | `~/.cline/` |
| Env Override | `CLINE_DIR` |
| Credential Files | `data/secrets.json` |
| Credential Notes | `data/secrets.json` holds API keys (encrypted at rest). Created via `cline auth -p <provider> -k <key>`. |
| Settings Files | `data/globalState.json`, `data/settings/cline_mcp_settings.json` |
| State Files | `data/workspace/`, `data/tasks/`, `log/` |
| Platform | All platforms: `~/.cline/`. Windows: `%USERPROFILE%\.cline\`. macOS VS Code extension uses `~/Library/Application Support/Code/User/globalStorage/saoudrizwan.claude-dev/`. |

---

## 12. Continue CLI

| Field | Value |
|-------|-------|
| ID | `continue` |
| Config Dirs | `~/.continue/` |
| Env Override | `--config <path>` flag |
| Credential Files | `config.yaml` (API keys inline under provider fields) |
| Credential Notes | API keys are stored directly in `config.yaml` under provider config sections. No separate secrets file. Can also use env vars. |
| Settings Files | `config.yaml` (primary), `config.json` (legacy), `config.ts` (programmatic override) |
| State Files | Managed by IDE extension or `cn` daemon |
| Platform | All platforms: `~/.continue/`. Windows: `%USERPROFILE%\.continue\`. |

---

## 13. Amp (Sourcegraph)

| Field | Value |
|-------|-------|
| ID | `amp` |
| Config Dirs | `~/.config/amp/` |
| Env Override | `--settings-file <path>` flag |
| Credential Files | Web-based auth (ampcode.com login) — token storage location not publicly documented |
| Credential Notes | Authentication is web-based via ampcode.com. Tokens likely stored in OS keychain or within `~/.config/amp/`. No standalone credential file documented. |
| Settings Files | `settings.json` or `settings.jsonc` |
| State Files | Not publicly documented |
| Extension Files | `skills/`, `AGENTS.md` |
| Platform | Linux/macOS: `~/.config/amp/`. Windows: `%USERPROFILE%\.config\amp\`. Enterprise: `/etc/ampcode/managed-settings.json` (Linux), `/Library/Application Support/ampcode/managed-settings.json` (macOS). |

---

## 14. Goose (Block)

| Field | Value |
|-------|-------|
| ID | `goose` |
| Config Dirs | `~/.config/block/goose/`, `~/.local/share/goose/`, `~/.local/state/goose/` |
| Env Override | none documented |
| Credential Files | OS keyring (macOS Keychain, Linux Secret Service, Windows Credential Manager) |
| Credential Notes | Credentials stored in **OS keyring by default**. Disable with `GOOSE_DISABLE_KEYRING=1` env var — falls back to local YAML file. API keys can also be set via env vars (OPENAI_API_KEY, etc.). NOT stored in config.yaml. |
| Settings Files | `~/.config/block/goose/config.yaml` |
| State Files | `~/.local/share/goose/sessions/`, `~/.local/state/goose/logs/` |
| Platform | Linux/macOS: `~/.config/block/goose/`. Windows: `%APPDATA%\Block\goose\config\`. |

---

## 15. Amazon Q Developer CLI

| Field | Value |
|-------|-------|
| ID | `amazonq` |
| Config Dirs | `~/.aws/amazonq/` |
| Env Override | `AWS_CONFIG_FILE`, `AWS_SHARED_CREDENTIALS_FILE` (for shared AWS files) |
| Credential Files | `~/.aws/credentials`, `~/.aws/config` (shared AWS credential chain) |
| Credential Notes | Uses standard AWS credential chain (`~/.aws/credentials` + `~/.aws/config`). Login via `q login`. Supports Builder ID (free) and Pro (IAM Identity Center). **IMPORTANT**: `~/.aws/credentials` is shared with all AWS tools — do NOT move this file, only reference it. |
| Settings Files | `default.json`, `mcp.json` (legacy) |
| State Files | `cache/`, `history/` |
| Extension Files | `cli-agents/`, `prompts/`, `profiles/` |
| Platform | All platforms: `~/.aws/amazonq/`. Windows: `%USERPROFILE%\.aws\amazonq\`. |
| Special Note | **Do NOT symlink `~/.aws/` itself** — only `~/.aws/amazonq/`. The `~/.aws/credentials` and `~/.aws/config` files are shared by all AWS tools and must not be moved. |

---

## 16. Kiro CLI (AWS)

| Field | Value |
|-------|-------|
| ID | `kiro` |
| Config Dirs | `~/.kiro/` |
| Env Override | none documented |
| Credential Files | Uses AWS credential chain (`~/.aws/credentials`, `~/.aws/config`) |
| Credential Notes | Same AWS credential chain as Amazon Q. No Kiro-specific credential file. |
| Settings Files | `settings/mcp.json`, `settings/` directory |
| State Files | Not publicly documented |
| Platform | All platforms: `~/.kiro/`. |

---

## 17. Kilo Code CLI

| Field | Value |
|-------|-------|
| ID | `kilo` |
| Config Dirs | `~/.config/kilo/` |
| Env Override | none documented |
| Credential Files | Managed via `kilo auth` — stored internally |
| Credential Notes | API keys set via `kilo auth` command or as env vars. Can also set inline in `opencode.json` under `provider.<name>.options.apiKey`. |
| Settings Files | `opencode.json` or `opencode.jsonc`, `config.json` |
| State Files | Not publicly documented |
| Platform | All platforms: `~/.config/kilo/`. |

---

## Summary: Credential Migration Table

Files that MUST be copied when creating vanilla profiles to avoid re-authentication:

| Tool | Credential File(s) | Shared? |
|------|-------------------|---------|
| Claude | `~/.claude/.credentials.json` | No |
| Codex | `~/.codex/auth.json` | No |
| Gemini | `~/.gemini/oauth_creds.json`, `~/.gemini/google_accounts.json`, `~/.gemini/mcp-oauth-tokens.json` | No |
| OpenCode | `~/.local/share/opencode/auth.json` | No |
| Pi | `~/.pi/agent/auth.json` | No |
| Copilot | `~/.config/github-copilot/apps.json` | No |
| OMC/OMX/OMG | none | — |
| Aider | `~/.aider.env` | No |
| Cline | `~/.cline/data/secrets.json` | No |
| Continue | `~/.continue/config.yaml` (keys inline — extract carefully) | No |
| Amp | Web-based auth (unknown file) | — |
| Goose | OS keyring (not file-based by default) | Shared |
| Amazon Q | `~/.aws/credentials` + `~/.aws/config` | **Shared with all AWS tools** |
| Kiro | `~/.aws/credentials` + `~/.aws/config` | **Shared with all AWS tools** |
| Kilo | Internal auth store | No |

### Special Cases for `raise`

1. **Aider dotfiles**: Not a directory — individual dotfiles in `~/`. Must be managed as a set of files, not a symlinked directory.
2. **OpenCode multi-dir**: 4 separate XDG directories. All must be swapped atomically.
3. **Copilot dual-dir**: 2 separate directories with different purposes (auth vs config).
4. **AWS shared credentials**: `~/.aws/credentials` and `~/.aws/config` are shared by Amazon Q, Kiro, and all other AWS CLI tools. `raise` should **never move or symlink these files**. Only manage `~/.aws/amazonq/` and `~/.kiro/`.
5. **Goose keyring**: Credentials in OS keyring can't be file-swapped. `raise` should document this limitation and offer `GOOSE_DISABLE_KEYRING=1` as workaround.
6. **Continue inline keys**: API keys embedded in `config.yaml` — no separate credential file. Migration means copying the entire config.
7. **Amp unknown auth**: Auth token storage not documented. `raise` should handle gracefully if auth files are discovered.

### Directory Count per Tool

| Tool | Dirs to Manage |
|------|---------------|
| Claude | 1 |
| Codex | 1 |
| Gemini | 1 |
| OpenCode | 4 |
| Pi | 1 |
| Copilot | 2 |
| OMC | 1 |
| OMX | 1 |
| OMG | 1 |
| Aider | 1 dir + 4 dotfiles |
| Cline | 1 |
| Continue | 1 |
| Amp | 1 |
| Goose | 3 |
| Amazon Q | 1 (only `~/.aws/amazonq/`, NOT `~/.aws/`) |
| Kiro | 1 |
| Kilo | 1 |
| **Total** | **~23 directories/file-sets** |
