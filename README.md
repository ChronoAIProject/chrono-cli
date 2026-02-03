# Chrono CLI

Command-line tool for the Developer Platform - deploy applications with CI/CD pipelines.

## Installation

### One-line install (macOS/Linux)

```bash
curl -sSL https://raw.githubusercontent.com/ChronoAIProject/chrono-cli/main/install.sh | sh
```

### Manual download

```bash
# macOS (Apple Silicon)
curl -sSL -o chrono https://github.com/ChronoAIProject/chrono-cli/releases/latest/download/chrono-darwin-arm64
chmod +x chrono && sudo mv chrono /usr/local/bin/

# macOS (Intel)
curl -sSL -o chrono https://github.com/ChronoAIProject/chrono-cli/releases/latest/download/chrono-darwin-amd64
chmod +x chrono && sudo mv chrono /usr/local/bin/

# Linux (x64)
curl -sSL -o chrono https://github.com/ChronoAIProject/chrono-cli/releases/latest/download/chrono-linux-amd64
chmod +x chrono && sudo mv chrono /usr/local/bin/

# Linux (ARM64)
curl -sSL -o chrono https://github.com/ChronoAIProject/chrono-cli/releases/latest/download/chrono-linux-arm64
chmod +x chrono && sudo mv chrono /usr/local/bin/
```

All releases: https://github.com/ChronoAIProject/chrono-cli/releases

## Development

### Prerequisites

- Go 1.22 or later

### Build from source

```bash
# Clone the repository
git clone https://github.com/ChronoAIProject/chrono-cli.git
cd chrono-cli

# Build
make build

# Or build directly
go build -o chrono .
```

### Install locally

```bash
# Install to ~/go/bin
make install-local

# Install system-wide (requires sudo)
make install
```

### Run tests

```bash
# Unit tests
make test

# Integration tests
make test-integration
```

### Build release binaries

```bash
make release
```

This builds binaries for all platforms (darwin/linux/amd64/arm64).

## Usage

```bash
# Initialize project (copies skills to .chrono/skills/)
chrono init

# Login to the platform
chrono login

# Detect project type
chrono detect

# Show version
chrono version

# Show help
chrono --help
```

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
