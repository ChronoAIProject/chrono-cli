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

**Example for Go:**
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

**Example for Node.js:**
```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
EXPOSE 8080
CMD ["npm", "start"]
```

**Example for Python:**
```dockerfile
FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE 8080
CMD ["python", "app.py"]
```

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
- [ ] Frontend is SPA
- [ ] Both have entry points

### Backend-Only Projects:
- [ ] `Dockerfile` exists in root
- [ ] Entry point exists in root (main.go, package.json with "start", app.py)

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

# For backend-only projects:
test -f Dockerfile && echo "✓ Dockerfile exists (REQUIRED)" || echo "✗ FAIL: Missing Dockerfile"
ls main.go package.json requirements.txt 2>/dev/null && echo "✓ Entry point found" || echo "✗ FAIL: No entry point found"

# For frontend-only projects:
grep -q '"build"' package.json && echo "✓ Build script exists (REQUIRED)" || echo "✗ FAIL: Missing build script"
```

**✅ PASS CRITERIA:**
- Backend projects: Dockerfile exists at required location
- Frontend projects: Build script exists in package.json
- Fullstack projects: Both backend/Dockerfile and frontend build script exist

**❌ FAIL if:**
- Missing Dockerfile for backend/fullstack projects
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
