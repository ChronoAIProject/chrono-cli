# Agent Instructions

## Skills
Custom Chrono skills are located at:

.chrono/skills/

Each skill is a markdown file that provides guidance for AI agents.
AI agents should inspect this directory before taking action.

## Available Skills

- **check-setup.md** - Verify project structure before deployment
- **deploy.md** - Deploy current project to the platform
- **restart.md** - Rolling restart backend deployment
- **agents.md** - This file - instructions for AI agents

## Preferred Behavior

- **Reuse existing skills** instead of re-implementing logic
- **Read skill files** before taking action to understand the proper workflow
- **Ask before creating new skills** - the existing skills cover most use cases
- **Always read .env files** before deployment to detect configuration
- **Always confirm with user** before creating pipelines or deploying

## Skill Usage

When a user asks you to deploy, restart, or check setup:
1. Read the relevant skill file from .chrono/skills/
2. Follow the steps outlined in the skill
3. Do not skip steps or make assumptions
4. If something is unclear, ask the user

## Critical: Environment Variables

**Before creating a pipeline, you MUST:**

1. **Read all .env files** in the project:
   - Root: `.env`, `.env.local`, `.env.production`
   - Frontend: `frontend/.env*`
   - Backend: `backend/.env*`

2. **Classify variables:**
   - `frontendEnvVars` - public prefixes (`NEXT_PUBLIC_*`, `VITE_*`, etc.) - baked into bundle
   - `backendEnvVars` - non-sensitive backend config (PORT, LOG_LEVEL, etc.)
   - `secrets` - sensitive data (API_KEY, JWT_SECRET, passwords) - stored securely

3. **Present detected variables** to the user in a clear table format

4. **Ask user to confirm or modify** before proceeding:
   - "Are these environment variables correct?"
   - "Do you need to add any additional variables?"
   - "Please provide values for any secrets (API keys, etc.)"
   - "Should I provision any databases (MongoDB, Redis, PostgreSQL)?"

5. **Wait for user confirmation** - never auto-create pipelines without consent

**Note:** Backend env vars are injected via Kubernetes secrets at runtime (secure).

## MCP Tools

The platform provides MCP tools for direct integration:

**GitHub & Projects:**
- `check_github_connection` - Check GitHub OAuth status
- `list_projects` - List all projects
- `create_project` - Create a new project

**Pipelines:**
- `list_pipelines` - List all pipelines
- `create_pipeline` - Create CI/CD pipeline (supports frontendEnvVars, backendEnvVars, and secrets)
- `trigger_pipeline_run` - Trigger deployment
- `get_run_status` - Check deployment status

**Deployments:**
- `restart_deployment` - Rolling restart
- `get_deployment_status` - Get deployment status

Prefer using MCP tools over shell commands when available.
