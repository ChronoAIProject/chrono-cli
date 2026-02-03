# Dockerfile Examples

**Important:** Build context is the **repository root**, not the Dockerfile's directory.
All `COPY` paths must be relative to repository root.

## Go

### Backend-Only (Dockerfile in root)
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

### Fullstack (Dockerfile in backend/)
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

## Node.js

### Backend-Only (Dockerfile in root)
```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
EXPOSE 8080
CMD ["npm", "start"]
```

### Fullstack (Dockerfile in backend/)
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

## Python

### Backend-Only (Dockerfile in root)
```dockerfile
FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE 8080
CMD ["python", "app.py"]
```

### Fullstack (Dockerfile in backend/)
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

## Common Mistakes

| Wrong | Correct | Reason |
|-------|---------|--------|
| `COPY . .` | `COPY backend/ ./` | Copies entire repo including frontend |
| `COPY package*.json ./` | `COPY backend/package*.json ./` | Copies root files, not backend files |
| `COPY requirements.txt .` | `COPY backend/requirements.txt ./` | Copies root files, not backend files |
| `COPY go.mod .` | `COPY backend/ ./` | Copies root files, not backend files |

**Always use paths relative to REPO ROOT in COPY commands.**
