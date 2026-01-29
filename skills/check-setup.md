# Check Project Setup

Verify your project structure meets platform requirements before deployment.

## Step 1: Detect Project Type

**Check which folders exist:**

```bash
ls -la
```

**Check for both backend and frontend folders:**
```bash
test -d backend && echo "✓ backend/" || echo "✗ no backend/"
test -d frontend && echo "✓ frontend/" || echo "✗ no frontend/"
```

---

## Detected Project Type

### 🟣 FULLSTACK (both backend/ AND frontend/ exist)

Use monorepo structure:
```
project/
├── backend/
│   ├── Dockerfile    ← REQUIRED
│   └── (source code)
└── frontend/
    └── (SPA code)
```

→ Complete **both Backend and Frontend checks** below.

---

### 🔵 BACKEND-ONLY (no backend/ or frontend/ folders, Dockerfile in root)

Use root folder structure:
```
project/
├── Dockerfile        ← REQUIRED (in root)
├── go.mod            (or package.json, requirements.txt)
└── (source code)
```

→ Skip to **Backend-only check** below.

---

### 🟢 FRONTEND-ONLY (no backend/ or frontend/ folders, package.json in root)

Use root folder structure:
```
project/
├── package.json
├── index.html        (or auto-generated)
└── (SPA source code)
```

→ Skip to **Frontend-only check** below.

---

## Backend Checks (Fullstack)

For fullstack projects with `backend/` folder:

### Required: Dockerfile

```bash
ls -la backend/Dockerfile
```

**❌ If missing:** Create `backend/Dockerfile`

**Example for Go (Backend-Only - Dockerfile in root):**
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest
COPY --from=builder /app/server /server
EXPOSE 8080
CMD ["/server"]
```

**Example for Go (Fullstack/Monorepo - Dockerfile in backend/):**
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app

# IMPORTANT: Build context is REPO ROOT, not backend/ folder
# Copy only backend/ subdirectory to avoid copying frontend/
COPY backend/ ./
RUN go build -o server ./cmd/server

FROM alpine:latest
COPY --from=builder /app/server /server
EXPOSE 8080
CMD ["/server"]
```

**Example for Node.js (Backend-Only - Dockerfile in root):**
```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
EXPOSE 8080
CMD ["npm", "start"]
```

**Example for Node.js (Fullstack/Monorepo - Dockerfile in backend/):**
```dockerfile
FROM node:20-alpine
WORKDIR /app

# IMPORTANT: Build context is REPO ROOT, not backend/ folder
# All COPY paths must be relative to repository root
COPY backend/package*.json ./
RUN npm ci --only=production

# Copy backend source files from backend/ subdirectory
COPY backend/ ./

# For TypeScript: use the built file as entrypoint
# CMD ["node", "dist/index.js"]

# Or use npm start (ensure backend/package.json has: "start": "node dist/index.js")
CMD ["npm", "start"]
```

**Example for Python (Backend-Only - Dockerfile in root):**
```dockerfile
FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE 8080
CMD ["python", "app.py"]
```

**Example for Python (Fullstack/Monorepo - Dockerfile in backend/):**
```dockerfile
FROM python:3.12-slim
WORKDIR /app

# IMPORTANT: Build context is REPO ROOT, not backend/ folder
# All COPY paths must be relative to repository root
COPY backend/requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

# Copy only backend/ subdirectory
COPY backend/ ./
EXPOSE 8080
CMD ["python", "app.py"]
```

### Required: Health Check Endpoint

**Your backend MUST implement a `/health` endpoint for Kubernetes readiness/liveness probes.**

```bash
# Verify health endpoint exists
grep -r "/health" backend/
```

**Health Check Requirements:**
| Setting | Value |
|---------|-------|
| **Default path** | `/health` (configurable via `healthCheckPath` in pipeline config) |
| **Success codes** | HTTP 200-399 (2xx = healthy, 3xx = redirect) |
| **Failure codes** | HTTP 400-599 (4xx/5xx = not ready/unhealthy) |
| **Timeout** | 10 seconds per request |
| **Poll interval** | Every 10 seconds |

**Expected behavior:**
- Return quickly (< 5 seconds recommended, < 10 seconds required)
- Check critical dependencies (database, external services)
- Return HTTP 200 with simple response like `{"status": "ok"}`

**Example implementations:**

**Node.js (Express):**
```javascript
app.get('/health', (req, res) => {
  res.status(200).json({ status: 'ok' });
});
```

**Node.js (Fastify):**
```javascript
fastify.get('/health', async (request, reply) => {
  return { status: 'ok' };
});
```

**Go (net/http):**
```go
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ok"}`))
})
```

**Python (Flask):**
```python
@app.route('/health')
def health():
    return {'status': 'ok'}, 200
```

**Python (FastAPI):**
```python
@app.get('/health')
def health():
    return {'status': 'ok'}
```

**❌ If missing:** Add a `/health` endpoint to your backend before deploying.

---

## Backend-Only Check

For backend projects using root folder:

### Required: Dockerfile in root

```bash
ls -la Dockerfile
```

**❌ If missing:** Create `Dockerfile` in root (use examples above)

### Verify entry point in root

```bash
# Go
ls main.go

# Node.js
grep '"start"' package.json

# Python
ls app.py main.py 2>/dev/null
```

### Required: Health Check Endpoint

**Your backend MUST implement a `/health` endpoint for Kubernetes readiness/liveness probes.**

```bash
# Verify health endpoint exists
grep -r "/health" .
```

**Health Check Requirements:**
| Setting | Value |
|---------|-------|
| **Default path** | `/health` (configurable via `healthCheckPath` in pipeline config) |
| **Success codes** | HTTP 200-399 (2xx = healthy, 3xx = redirect) |
| **Failure codes** | HTTP 400-599 (4xx/5xx = not ready/unhealthy) |
| **Timeout** | 10 seconds per request |
| **Poll interval** | Every 10 seconds |

**Expected behavior:**
- Return quickly (< 5 seconds recommended, < 10 seconds required)
- Check critical dependencies (database, external services)
- Return HTTP 200 with simple response like `{"status": "ok"}`

See example implementations in the **Backend Checks (Fullstack)** section above.

**❌ If missing:** Add a `/health` endpoint to your backend before deploying.

---

## Frontend Checks (Fullstack)

For fullstack projects with `frontend/` folder:

### Required: Must be SPA (Single Page Application)

**Check framework:**
```bash
cat frontend/package.json
```

**✅ Valid SPA frameworks:**
- Next.js (use `output: 'export'` for static)
- React (CRA, Vite)
- Vue.js
- Angular
- Svelte

**❌ NOT supported:**
- Server-side rendering (SSR) without static export
- Static site generators (Hugo, Jekyll - use for docs only)
- Multi-page applications (MPA)

**For Next.js users:**
Ensure static export in `next.config.js`:
```js
/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  images: { unoptimized: true }
}
module.exports = nextConfig
```

### Verify Build Script

```bash
grep -A 3 '"build"' frontend/package.json
```

---

## Frontend-Only Check

For frontend projects using root folder:

### Required: Must be SPA

**Check framework:**
```bash
cat package.json
```

**✅ Valid SPA frameworks:**
- Next.js (static export)
- React (CRA, Vite)
- Vue.js
- Angular
- Svelte

### Verify Build Script

```bash
grep -A 3 '"build"' package.json
```

---

## Common Issues

### Issue: "Docker build fails" - file not found, wrong files copied, or start command fails

**Cause:** Docker build context is the **repository root**, not the Dockerfile's directory.

**This affects ALL languages** (Node.js, Go, Python, etc.) when using monorepo structure.

**For monorepo structure (`backend/Dockerfile`):**
```
repo/                    ← Build context root (COPY paths are relative here)
  go.mod                 ← Root go.mod
  requirements.txt       ← Root requirements.txt
  package.json           ← Root package.json (no start script)
  backend/
    Dockerfile           ← Dockerfile location
    go.mod               ← Backend go.mod ✓
    requirements.txt     ← Backend requirements.txt ✓
    package.json         ← Backend package.json (has start script) ✓
    main.go / app.py     ← Backend entry point ✓
```

**Wrong Dockerfile (affects all languages):**
```dockerfile
COPY . .                              # Copies ENTIRE repo (including frontend/) ✗
COPY package*.json ./                 # Copies ROOT files ✗
COPY requirements.txt .               # Copies ROOT files ✗
COPY go.mod .                         # Copies ROOT files ✗
```

**Correct Dockerfile examples:**

| Language | Wrong | Correct |
|----------|-------|---------|
| Node.js | `COPY package*.json ./` | `COPY backend/package*.json ./` |
| Python | `COPY requirements.txt .` | `COPY backend/requirements.txt ./` |
| Go | `COPY . .` | `COPY backend/ ./` |
| All | `COPY . .` | `COPY backend/ ./` |

**Always use paths relative to REPO ROOT in COPY commands.**

---

### Issue: "I want fullstack but my code is in root"

**Solution - Reorganize into monorepo:**
```bash
mkdir backend frontend
mv (backend files) backend/
mv (frontend files) frontend/
# Create backend/Dockerfile
```

### Issue: "Dockerfile missing"

**For fullstack:** Create `backend/Dockerfile`

**For backend-only:** Create `Dockerfile` in root

Use examples above based on your language.

### Issue: "Frontend is not an SPA"

**Convert to SPA:**
- Enable client-side routing
- Disable SSR
- Build static assets (HTML, JS, CSS)
- All API calls to backend (not server-rendered)

---

## Quick Checklist

### Fullstack Projects:
- [ ] `backend/` folder exists
- [ ] `frontend/` folder exists
- [ ] `backend/Dockerfile` exists
- [ ] Backend has `/health` endpoint (check: `grep -r "/health" backend/`)
- [ ] Frontend is SPA
- [ ] Both have entry points
- [ ] `backend/Dockerfile` uses correct COPY paths (check: `grep "^COPY" backend/Dockerfile`)

### Backend-Only Projects:
- [ ] `Dockerfile` exists in root
- [ ] Entry point exists in root (main.go, package.json with "start", app.py)
- [ ] Backend has `/health` endpoint (check: `grep -r "/health" .`)

### Frontend-Only Projects:
- [ ] `package.json` exists in root
- [ ] Frontend is SPA (React, Next.js, Vue, Angular)
- [ ] Has "build" script

### All Projects:
- [ ] .chrono/config.yaml exists (run "chrono init")

---

## Step 2: Run Detection (Optional)

After fixing structure, verify with:

```bash
chrono detect --save
```

This creates `.chrono/metadata.yaml` with detected configuration.

---

## Final Verification

**Before declaring setup as PASS, verify:**

```bash
# For fullstack projects:
test -d backend && echo "✓ backend/ exists"
test -d frontend && echo "✓ frontend/ exists"
test -f backend/Dockerfile && echo "✓ backend/Dockerfile exists (REQUIRED)" || echo "✗ FAIL: Missing backend/Dockerfile"
grep -q '"build"' frontend/package.json && echo "✓ Frontend has build script" || echo "✗ FAIL: Frontend missing build script"
grep -rq '"/health"' backend/ 2>/dev/null && echo "✓ Backend has /health endpoint (REQUIRED)" || echo "✗ FAIL: Backend missing /health endpoint"

# For backend-only projects:
test -f Dockerfile && echo "✓ Dockerfile exists (REQUIRED)" || echo "✗ FAIL: Missing Dockerfile"
ls main.go package.json requirements.txt 2>/dev/null && echo "✓ Entry point found" || echo "✗ FAIL: No entry point found"
grep -rq '"/health"' . 2>/dev/null && echo "✓ Backend has /health endpoint (REQUIRED)" || echo "✗ FAIL: Backend missing /health endpoint"

# For frontend-only projects:
grep -q '"build"' package.json && echo "✓ Build script exists (REQUIRED)" || echo "✗ FAIL: Missing build script"
```

**✅ PASS CRITERIA:**
- Backend projects: Dockerfile exists at required location **AND** `/health` endpoint implemented
- Frontend projects: Build script exists in package.json
- Fullstack projects: Both backend/Dockerfile and frontend build script exist **AND** backend has `/health` endpoint

**❌ FAIL if:**
- Missing Dockerfile for backend/fullstack projects
- Missing `/health` endpoint for backend/fullstack projects
- Missing build script for frontend projects
- Invalid project structure

**Report to user:**
- If PASS: "✅ Project structure verified successfully. Ready for deployment."
- If FAIL: "❌ Setup check failed. Please fix the issues above before deploying."

---

## Ready to Deploy?

Once all checks pass:
1. Run `chrono login`
2. Use `/deploy` skill in Cursor IDE
