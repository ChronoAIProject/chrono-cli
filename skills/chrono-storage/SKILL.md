---
name: chrono-storage
description: ChronoAI Platform S3 object storage integration for deployed apps. Use when implementing file upload functionality in apps deployed on ChronoAI via HTTP API. Covers upload API, CDN access patterns, error handling, and framework-specific implementations (React, Vue, vanilla JS).
---

# ChronoAI Storage API

## Quick Start

Upload a file and get the CDN URL:

```javascript
const formData = new FormData();
formData.append('file', fileBlob);

const response = await fetch('/api/v1/storage/pipelines/{pipelineId}/files', {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}` },
  body: formData
});

const { url, filename, size } = await response.json();
// url: "https://cdn.chrono-ai.fun/myapp/image.jpg"
```

## Core Concepts

1. **Upload via backend proxy** - POST multipart/form-data to `/api/v1/storage/pipelines/:pipelineId/files`
2. **CDN delivery** - Files served at `https://cdn.chrono-ai.fun/{appName}/{filename}`
3. **Env var** - `CHRONO_CDN_URL` injected automatically when Object Storage enabled

## Prerequisites

- Object Storage middleware enabled in pipeline config
- Authenticated session with API token

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/storage/pipelines/:pipelineId/files` | Upload file |
| GET | `/api/v1/storage/pipelines/:pipelineId/files` | List files |
| DELETE | `/api/v1/storage/pipelines/:pipelineId/files/:filename` | Delete file |

**Upload Response:**
```json
{
  "filename": "photo.jpg",
  "url": "https://cdn.chrono-ai.fun/myapp/photo.jpg",
  "size": 245632
}
```

**Limits:** Max 100MB per file

## Framework-Specific Implementations

- **React Hook:** See [react.md](references/react.md)
- **Vue Composable:** See [vue.md](references/vue.md)
- **Class-Based Utility:** See [class-based.md](references/class-based.md)

## CDN URL Pattern

```javascript
// CHRONO_CDN_URL = "https://cdn.chrono-ai.fun/{appName}"
const baseUrl = process.env.CHRONO_CDN_URL;
const fileUrl = `${baseUrl}/${filename}`;
```

## Error Handling

| Error | Cause | Handling |
|-------|-------|----------|
| "not enabled" | Object storage disabled for pipeline | Enable in pipeline config |
| "Unauthorized" | Invalid/expired token | Refresh authentication |
| "size exceeds" | File > 100MB | Show size limit to user |
