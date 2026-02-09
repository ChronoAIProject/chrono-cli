---
name: chrono-storage
description: ChronoAI Platform S3 object storage integration for deployed apps. Use when implementing file upload functionality in apps deployed on ChronoAI via HTTP API. Covers upload API, CDN access patterns, error handling, and framework-specific implementations (React, Vue, vanilla JS).
---

# chrono-storage

S3 object storage integration for apps deployed on the ChronoAI developer platform. Upload files via the platform HTTP API and access them through the integrated CDN.

## Quick Start

```javascript
// Upload a file and get the CDN URL
const formData = new FormData();
formData.append('file', fileInput.files[0]);

const response = await fetch(`${process.env.PLATFORM_URL}/api/v1/storage/pipelines/${process.env.PIPELINE_ID}/files`, {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${process.env.PLATFORM_API_TOKEN}`
  },
  body: formData
});

const { url, filename, size } = await response.json();
// Use: https://{CHRONO_CDN_URL}/{filename}
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `CHRONO_CDN_URL` | App's CDN base URL (includes app name) |
| `PIPELINE_ID` | Pipeline ID for API calls (24-char hex string) |
| `PLATFORM_URL` | Platform base URL for API endpoints |
| `PLATFORM_API_TOKEN` | Pipeline-scoped token for storage API |

## Reference Documentation

- [Environment Variables](references/environment-variables.md) - Detailed variable descriptions and security notes
- [Upload API](references/upload-api.md) - HTTP API endpoint for file uploads with code examples
- [CDN Access](references/cdn-access.md) - URL structure, caching, and image optimization
- [Error Handling](references/error-handling.md) - Common errors and troubleshooting
