---
name: agents
description: Meta-skill for AI agents working with Chrono CLI. Provides navigation to available skills and MCP tools. Use when starting work on Chrono-related tasks to understand available capabilities, or when unsure which skill to use for a specific task.
---

# Agent Instructions

## Available Skills

Chrono CLI provides these skills for AI agents:

- **check-setup** - Verify project structure before deployment
- **deploy** - Deploy current project to the platform
- **restart** - Rolling restart backend deployment
- **chrono** - Chrono CLI command reference

## Skill Selection Guide

| User Request | Use Skill |
|--------------|-----------|
| "Deploy my project" | deploy |
| "Is my project ready to deploy?" | check-setup |
| "Restart the pods" | restart |
| "How do I use chrono CLI?" | chrono |
| "Setup MCP integration" | chrono |

## Preferred Behavior

- **Reuse existing skills** instead of re-implementing logic
- **Read skill files** before taking action to understand the proper workflow
- **Ask before creating new skills** - the existing skills cover most use cases

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

4. **Ask user to confirm or modify** before proceeding

5. **Wait for user confirmation** - never auto-create pipelines without consent

**Note:** Backend env vars are injected via Kubernetes secrets at runtime (secure).

## MCP Tools

The platform provides MCP tools for direct integration:

**User:**
- `list_teams` - List all teams the user belongs to

**GitHub & Projects:**
- `check_github_connection` - Check GitHub OAuth status
- `start_github_device_flow` - Start GitHub device flow authorization
- `poll_github_device_flow` - Poll GitHub device flow status
- `list_github_repos` - List GitHub repositories
- `list_github_branches` - List branches for a GitHub repository
- `detect_repo_type` - Detect application type and structure
- `list_projects` - List all projects
- `create_project` - Create a new project
- `get_project` - Get details of a specific project

**Pipelines:**
- `list_pipelines` - List all pipelines (filtered by project/environment/status)
- `create_pipeline` - Create CI/CD pipeline (supports frontendEnvVars, backendEnvVars, secrets)
- `get_pipeline` - Get details of a specific pipeline including config
- `update_pipeline` - Update pipeline config (triggers auto-restart)

**Pipeline Runs:**
- `trigger_pipeline_run` - Trigger a new pipeline run
- `get_run_status` - Get current status of a pipeline run
- `list_runs` - List recent runs for a pipeline
- `cancel_run` - Cancel a running or pending pipeline run

**Deployments:**
- `restart_deployment` - Rolling restart
- `get_deployment_status` - Get deployment status
- `get_deployment_logs` - Get logs from pods (all pods with labels)
- `get_pod_env_vars` - Get env vars from pods (secrets masked)

Prefer using MCP tools over shell commands when available.
