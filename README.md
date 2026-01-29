# Chrono CLI

Command-line tool for the Developer Platform - deploy applications with CI/CD pipelines.

## Installation

```bash
# macOS (Apple Silicon)
curl -sSL -o chrono https://github.com/aelfdevops/developer-platform/releases/latest/download/chrono-darwin-arm64
chmod +x chrono && sudo mv chrono /usr/local/bin/

# macOS (Intel)
curl -sSL -o chrono https://github.com/aelfdevops/developer-platform/releases/latest/download/chrono-darwin-amd64
chmod +x chrono && sudo mv chrono /usr/local/bin/

# Linux (x64)
curl -sSL -o chrono https://github.com/aelfdevops/developer-platform/releases/latest/download/chrono-linux-amd64
chmod +x chrono && sudo mv chrono /usr/local/bin/

# Linux (ARM64)
curl -sSL -o chrono https://github.com/aelfdevops/developer-platform/releases/latest/download/chrono-linux-arm64
chmod +x chrono && sudo mv chrono /usr/local/bin/
```

All releases: https://github.com/aelfdevops/developer-platform/releases

## Skills

AI agent skills for deployment automation. These are used by Cursor and other AI coding assistants.

### Available Skills

| Skill | Description |
|-------|-------------|
| [deploy.md](skills/deploy.md) | Deploy current project to the platform |
| [restart.md](skills/restart.md) | Rolling restart backend deployment |
| [check-setup.md](skills/check-setup.md) | Verify project structure before deployment |
| [agents.md](skills/agents.md) | Instructions for AI agents |

### How Skills Work

When you run `chrono init` in your project, the skills are copied to `.chrono/skills/`. AI agents read these files to understand how to help you deploy.

### Using Skills in Cursor

Simply ask Cursor to deploy, restart, or check your project setup:

- "Deploy this project"
- "Restart the backend"
- "Check if my project is ready for deployment"

The AI will read the relevant skill and follow the steps.

## License

MIT
