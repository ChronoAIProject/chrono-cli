---
name: check-setup
description: Verify project structure meets platform requirements before deployment. Use when checking if a project is ready to deploy, verifying Dockerfile exists, ensuring health endpoints are implemented, or validating SPA frontend setup. Covers fullstack, backend-only, and frontend-only project types.
---

# Check Project Setup

Verify your project structure meets platform requirements before deployment.

## Quick Verification

```bash
# Check which folders exist
test -d backend && echo "✓ backend/" || echo "✗ no backend/"
test -d frontend && echo "✓ frontend/" || echo "✗ no frontend/"

# Quick checks
test -f backend/Dockerfile && echo "✓ Backend Dockerfile exists" || echo "✗ Missing backend/Dockerfile"
test -f Dockerfile && echo "✓ Root Dockerfile exists" || echo "✗ Missing root Dockerfile"
grep -q '"build"' frontend/package.json && echo "✓ Frontend has build script" || echo "✗ Frontend missing build script"
```

## Project Types

### Fullstack (both backend/ AND frontend/ exist)
- Requires `backend/Dockerfile` and `/health` endpoint
- Frontend must be SPA with build script

### Backend-Only (Dockerfile in root)
- Requires `Dockerfile` in root and `/health` endpoint

### Frontend-Only (package.json in root)
- Must be SPA (Next.js, React, Vue, Angular, Svelte)

## Required: Health Check Endpoint

**Your backend MUST implement a `/health` endpoint for Kubernetes readiness/liveness probes.**

| Setting | Value |
|---------|-------|
| **Default path** | `/health` (configurable via `healthCheckPath`) |
| **Success codes** | HTTP 200-399 |
| **Failure codes** | HTTP 400-599 |
| **Timeout** | 10 seconds |
| **Poll interval** | Every 10 seconds |

**Verify:** `grep -r "/health" backend/` or `grep -r "/health" .`

**Example implementations** - See [HEALTH_ENDPOINTS.md](references/HEALTH_ENDPOINTS.md) for code examples in Go, Node.js, Python, FastAPI, Flask.

## Required: Dockerfile

**For fullstack:** `backend/Dockerfile` required.
**For backend-only:** `Dockerfile` in root required.

**Important:** Build context is repository root. All `COPY` paths must be relative to repo root.

**Dockerfile examples** - See [DOCKERFILE_EXAMPLES.md](references/DOCKERFILE_EXAMPLES.md) for templates in Go, Node.js, Python.

## Quick Checklist

### Fullstack:
- [ ] `backend/` and `frontend/` exist
- [ ] `backend/Dockerfile` exists
- [ ] Backend has `/health` endpoint
- [ ] Frontend is SPA with build script

### Backend-Only:
- [ ] `Dockerfile` exists in root
- [ ] Entry point exists (main.go, package.json, app.py)
- [ ] Backend has `/health` endpoint

### Frontend-Only:
- [ ] `package.json` exists in root
- [ ] Frontend is SPA with build script

### All Projects:
- [ ] `.chrono/config.yaml` exists (run `chrono init`)

## Final Verification

```bash
# Fullstack
test -f backend/Dockerfile && echo "✓ PASS" || echo "✗ FAIL: Missing backend/Dockerfile"
grep -rq '"/health"' backend/ && echo "✓ PASS" || echo "✗ FAIL: Missing /health endpoint"

# Backend-only
test -f Dockerfile && echo "✓ PASS" || echo "✗ FAIL: Missing Dockerfile"
grep -rq '"/health"' . && echo "✓ PASS" || echo "✗ FAIL: Missing /health endpoint"

# Frontend-only
grep -q '"build"' package.json && echo "✓ PASS" || echo "✗ FAIL: Missing build script"
```

**✅ PASS:** Project verified successfully. Ready for deployment.
**❌ FAIL:** Fix issues before deploying.

## Optional: Save Detection

After fixing structure:
```bash
chrono detect --save
```
Creates `.chrono/metadata.yaml` with detected configuration.
