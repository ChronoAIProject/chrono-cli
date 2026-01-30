# Chrono CLI

Chrono CLI allows you to interact with the Developer Platform without accessing the dashboard.

## Installation

Install globally with:
```bash
go build -o ~/.local/bin/chrono .
```

## Available Commands

### Authentication

**Login** - Authenticate via Keycloak device flow
```bash
chrono login
```

**Logout** - Clear local credentials
```bash
chrono logout
```

**Status** - Show current login status
```bash
chrono status
```

### Project Setup

**Init** - Initialize Chrono configuration for your project
```bash
chrono init
```
Creates `.chrono/config.yaml` with project settings.

**Detect** - Analyze project and detect tech stack
```bash
chrono detect [--save]
```
- `--save` - Save detected configuration as metadata

### AI Editor Integration

**MCP Setup** - Configure AI editor to use Chrono as an MCP server
```bash
chrono mcp-setup [editor]
```

Supported editors:
- `cursor` (default) - Cursor IDE
- `claude-code` - Claude Code
- `codex` - Codex
- `gemini` - Gemini CLI

This command:
1. Creates API token automatically
2. Configures editor's MCP settings
3. Verifies MCP connection

## Configuration Files

```
~/.chrono/config.yaml     # Main CLI config
.chrono/config.yaml       # Project-specific config
.chrono/metadata.yaml      # Auto-generated project metadata
.cursor/mcp.json          # Cursor MCP config (auto-generated)
.mcp.json                 # Claude Code MCP config (auto-generated)
```

## Common Workflows

### First-time setup for a new project:
```bash
1. chrono login
2. cd /path/to/project
3. chrono init
4. chrono detect --save
5. chrono mcp-setup
```

### Deploy current project:
```bash
# Use the MCP tools via your AI editor
- list_projects
- create_pipeline
- trigger_pipeline_run
- get_run_status
```

## Global Flags

- `--api-url string` - API server URL (overrides config)
- `--config string` - Config file path
- `--debug` - Enable debug output
- `-h, --help` - Help for any command
- `-v, --version` - Show version