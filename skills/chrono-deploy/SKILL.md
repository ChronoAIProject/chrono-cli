---
name: chrono-deploy
description: Deploy current Git repository to the Developer Platform using MCP tools. Use when user asks to deploy, to deploy their project, or wants to push code to production. Covers pre-flight checks, GitHub OAuth, pipeline creation with environment variable detection, deployment triggering, and monitoring.
---

# Deploy Current Project

Deploy the current Git repository using developer-platform MCP.

## Pre-Flight Check: Verify Project Structure

**IMPORTANT:** Before deploying, you MUST verify the project structure is correct.

**First, follow the check-setup skill:**

1. Read the check-setup skill to understand the verification steps
2. Check if `backend/Dockerfile` exists (for backend/fullstack projects)
3. Verify frontend is a valid SPA (for frontend/fullstack projects)
4. Ensure all required files are in place

```bash
# Quick verification checks:
# For fullstack/backend projects:
test -f backend/Dockerfile && echo "✓ Backend Dockerfile exists" || echo "✗ Missing backend/Dockerfile"

# For frontend projects:
grep -q '"build"' frontend/package.json 2>/dev/null && echo "✓ Frontend has build script" || echo "✗ Frontend missing build script"
grep -q '"build"' package.json 2>/dev/null && echo "✓ Root has build script" || echo "✗ Root missing build script"
```

**❌ If checks fail:**
- Follow the check-setup skill to fix the structure
- Create the required Dockerfile for backend projects
- Ensure build scripts exist for frontend projects
- Only proceed with deployment once all checks pass

**✅ If all checks pass:** Continue with deployment below.

---

## Step 0: Check for Project Metadata

**First, check if .chrono/metadata.yaml exists:**

```bash
if [ -f .chrono/metadata.yaml ]; then
    echo "✓ Found project metadata"
    cat .chrono/metadata.yaml
else
    echo "⚠ No metadata found - will detect project structure"
fi
```

**If metadata exists:**
- Use the detected project type, tech stack, and middleware directly
- Skip to Step 3 with pre-populated values
- The metadata contains: project type, frameworks, ports, dockerfile paths, env vars

**If no metadata:**
- Continue with detection steps below
- Consider running "chrono detect --save" to create metadata for faster future deployments

## Step 1: Check Git Status

```bash
git config --get remote.origin.url
git branch --show-current
git status --porcelain
git log origin/$(git branch --show-current)..HEAD --oneline 2>/dev/null
```

Parse repo as owner/repo from the URL.

### Pre-flight:
- **Uncommitted changes?** → Ask to commit first
- **Unpushed commits?** → Ask to push first, show the commits
- **Clean?** → Continue

## Step 2: Check GitHub Connection

Use `check_github_connection`. If not connected:

1. Use `start_github_device_flow` to get the userCode and deviceCode
2. **Open browser** by running:
   ```bash
   open "https://github.com/login/device"
   ```
3. Display the code prominently to the user like this:

   ---
   **GitHub Authorization**

   I've opened GitHub in your browser.

   Enter this code: **`XXXX-XXXX`**

   Then click **Continue** → **Authorize**

   ---

4. Use `poll_github_device_flow` with the deviceCode until success (poll every 5 seconds)

## Step 3: Find or Create Pipeline

Use `list_pipelines` to find existing pipeline for this repo+branch.

**If exists:** Use it.
**If not:** Create a new pipeline:

### 3a. Get Project Information

**Option A: From .chrono/metadata.yaml (if exists)**
- Read project type (frontend/backend/fullstack)
- Use detected frameworks, ports, and middleware
- Use detected environment variables
- Skip detection steps

**Option B: Detect Project (if no metadata)**

1. `detect_repo_type` to get app type (frontend/backend/fullstack)
2. `list_projects` / `create_project`
3. **Analyze code for middleware dependencies:**

   Before asking the user, scan the codebase to detect middleware:

   **For Node.js (check package.json dependencies):**
   - `mongodb`, `mongoose`, `@prisma/client` (with mongodb) → suggest MongoDB
   - `redis`, `ioredis`, `@redis/client` → suggest Redis
   - `pg`, `postgres`, `@prisma/client` (with postgresql) → suggest PostgreSQL

   **For Python (check requirements.txt or pyproject.toml):**
   - `pymongo`, `motor`, `mongoengine` → suggest MongoDB
   - `redis`, `aioredis` → suggest Redis
   - `psycopg2`, `asyncpg`, `sqlalchemy` → suggest PostgreSQL

   **For Go (check go.mod):**
   - `go.mongodb.org/mongo-driver` → suggest MongoDB
   - `github.com/redis/go-redis`, `github.com/go-redis/redis` → suggest Redis
   - `github.com/lib/pq`, `github.com/jackc/pgx` → suggest PostgreSQL

   **For .NET/C# (check *.csproj PackageReference or packages.config):**
   - `MongoDB.Driver` → suggest MongoDB
   - `StackExchange.Redis`, `Microsoft.Extensions.Caching.StackExchangeRedis` → suggest Redis
   - `Npgsql`, `Npgsql.EntityFrameworkCore.PostgreSQL` → suggest PostgreSQL

4. **Read and parse .env files:**

   **REQUIRED:** You MUST read .env files to detect environment variables.

   **Search for .env files in this order:**

   ```bash
   # Check which .env files exist
   ls -la .env* 2>/dev/null
   ls -la frontend/.env* 2>/dev/null
   ls -la backend/.env* 2>/dev/null
   ```

   **Read files in priority order (first found wins for each variable):**

   For **Frontend** (check `frontend/` first, then root):
   - `.env.production` → production values (preferred)
   - `.env.local` → local overrides
   - `.env` → default values

   For **Backend** (check `backend/` first, then root):
   - `.env.production`
   - `.env.local`
   - `.env`

   **Parse each .env file:**
   ```bash
   # Read and display env vars (excluding comments and empty lines)
   grep -v '^#' .env | grep -v '^$' | grep '='
   ```

   **Identify framework-specific prefixes:**
   - `NEXT_PUBLIC_*` → Next.js (exposed to browser)
   - `REACT_APP_*` → Create React App (exposed to browser)
   - `VITE_*` → Vite (exposed to browser)
   - `VUE_APP_*` → Vue CLI (exposed to browser)
   - No prefix → Backend-only (keep server-side)

   **Classify variables:**

   **Frontend (public - baked into bundle):**
   - `NEXT_PUBLIC_*`, `REACT_APP_*`, `VITE_*`, `VUE_APP_*`
   - These are visible in browser, never put secrets here!
   - Goes in `frontendEnvVars`

   **Backend (non-sensitive config):**
   - ONLY these types: `PORT`, `LOG_LEVEL`, `NODE_ENV`, `HOST`, `DEBUG`
   - Simple configuration that isn't sensitive
   - Goes in `backendEnvVars`

   **Secrets (sensitive data - stored securely):**
   - **MUST go to secrets if name contains:** `KEY`, `SECRET`, `TOKEN`, `PASSWORD`, `CREDENTIAL`, `AUTH`, `PRIVATE`
   - Pattern match examples: `*_API_KEY`, `*_SECRET`, `*_TOKEN`, `JWT_*`, `*_PASSWORD`
   - Specific examples: `LLM_API_KEY`, `OPENAI_API_KEY`, `JWT_SECRET`, `DATABASE_PASSWORD`, `AUTH_TOKEN`
   - **ALWAYS use `secrets` field for these, NEVER `backendEnvVars`**
   - Goes in `secrets` field
   - Stored encrypted, injected at runtime via K8s secrets

   **⚠️ CRITICAL: When in doubt, put it in `secrets`. It's safer to over-protect than under-protect.**

5. **Present detected configuration and ASK user to confirm:**

   **You MUST show the user what you detected and get confirmation before creating the pipeline.**

   Format your message like this:

   ---
   **Detected Configuration**

   **Project Type:** {fullstack/backend/frontend}

   **Backend Port:** {detected port, default 3000}

   **Middleware Detected:**
   - {MongoDB/Redis/PostgreSQL or "None detected"}

   **Frontend Environment Variables:** (public, baked into bundle)
   | Variable | Value |
   |----------|-------|
   | NEXT_PUBLIC_API_URL | https://api.example.com |
   | NEXT_PUBLIC_APP_NAME | MyApp |

   **Backend Environment Variables:** (private, injected via K8s secrets)
   | Variable | Value |
   |----------|-------|
   | PORT | 3000 |
   | LOG_LEVEL | info |

   **Secrets:** (sensitive data, stored securely via K8s secrets)
   ⚠️ Any variable with KEY, SECRET, TOKEN, PASSWORD, AUTH in the name goes here!
   | Secret | Value |
   |--------|-------|
   | LLM_API_KEY | sk-xxx... |
   | OPENAI_API_KEY | sk-xxx... |
   | JWT_SECRET | (please provide) |
   | DATABASE_PASSWORD | (please provide) |

   ---
   **Please confirm or update:**
   1. Are these environment variables correct?
   2. Do you need to add any additional variables?
   3. Please provide values for the secrets listed above
   4. Should I provision MongoDB/Redis?

   ---

   **Wait for user response before proceeding.**

   **If user adds/modifies variables:**
   - Update your list with their changes
   - Confirm the final list before creating pipeline

   **If no .env files found:**
   - Ask: "I didn't find any .env files. Do you have environment variables to configure?"
   - For backend: "What port does your backend listen on?" (default: 3000)

   **For middleware:**
   - If detected: "I detected **{middleware}** dependencies. Should I provision these?"
   - If not detected: "Do you need a database or cache? (MongoDB, Redis, PostgreSQL)"
   - MongoDB → auto-injects `MONGODB_URI` and `MONGODB_DATABASE`
   - Redis → auto-injects `REDIS_URL`
   - PostgreSQL → auto-injects `DATABASE_URL`

6. `create_pipeline` with:
   - **name** and **appName**: Both use format {repo}-{branch-short}. Keep short and readable.
   - Detected settings from step 1
   - `frontendEnvVars`: public frontend variables (NEXT_PUBLIC_*, etc.)
   - `backendEnvVars`: **ONLY** non-sensitive config (PORT, LOG_LEVEL, NODE_ENV, HOST)
   - `secrets`: **ALL variables containing KEY, SECRET, TOKEN, PASSWORD, AUTH** (e.g., LLM_API_KEY, JWT_SECRET, DATABASE_PASSWORD)
   - `middleware`: `["mongodb"]`, `["redis"]`, or `["mongodb", "redis"]`
   - Backend port if not 8080

   **⚠️ IMPORTANT:** Double-check that no secrets ended up in `backendEnvVars`. Any *_KEY, *_SECRET, *_TOKEN MUST be in `secrets`.

7. **Show the user the generated URLs** from the response:
   - Frontend URL: https://{appName}.chrono-ai.fun
   - Backend URL: https://{appName}-api.chrono-ai.fun

## Step 4: Deploy

Use `trigger_pipeline_run` with pipelineId.

## Step 5: Monitor

Use `get_run_status` to poll progress. **Poll every 15 seconds** (builds take 2-5 minutes).

Report: stages, success URL, or failure details.

### Post-Deployment Verification (Optional)

After successful deployment, verify the application is running correctly:

```bash
# Check deployment health
get_deployment_status(pipelineId)

# Get pod logs for troubleshooting
get_deployment_logs(pipelineId)

# Verify env vars are set correctly (secrets masked)
get_pod_env_vars(pipelineId)
```

## Tip: Create Metadata for Faster Future Deployments

After successful deployment, create metadata to skip detection next time:

```bash
chrono detect --save
```

This creates .chrono/metadata.yaml with your project configuration.

## Important Notes

- **Deployments are manual only.** Use /deploy or trigger_pipeline_run each time you want to deploy new code.
- Automatic deployments on git push are NOT supported.
