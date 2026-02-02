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

## Skill Usage

When a user asks you to deploy, restart, or check setup:
1. Read the relevant skill file from .chrono/skills/ or via MCP tools
2. Follow the steps outlined in the skill
3. Do not skip steps or make assumptions
4. If something is unclear, ask the user

## MCP Tools

The platform also provides MCP tools for direct integration:
- check_github_connection - Check GitHub OAuth status
- list_projects - List all projects
- create_project - Create a new project
- list_pipelines - List all pipelines
- create_pipeline - Create CI/CD pipeline
- trigger_pipeline_run - Trigger deployment
- get_run_status - Check deployment status
- restart_deployment - Rolling restart
- get_deployment_status - Get deployment status

Prefer using MCP tools over shell commands when available.
