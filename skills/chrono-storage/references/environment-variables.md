# Environment Variables

## Auto-Injected Variables

When object storage is enabled for your pipeline, these environment variables are automatically injected into your backend container:

### CHRONO_CDN_URL

**Format:** `{CDN_BASE_URL}/{APP_NAME}`

Your app's CDN base URL for serving uploaded files. This URL includes your app name as a prefix.

```javascript
const cdnURL = process.env.CHRONO_CDN_URL;
// Example: "https://cdn.example.com/myapp"

// Construct file URLs
const fileURL = `${cdnURL}/${filename}`;
// Example: "https://cdn.example.com/myapp/resume.pdf"
```

### PIPELINE_ID

**Format:** ObjectId (24-character hex string)

Your pipeline's unique identifier. Use this when calling the platform's storage API.

```javascript
const pipelineID = process.env.PIPELINE_ID;
// Example: "507f1f77bcf86cd799439011"

// Use in API calls
const uploadURL = `${platformURL}/api/v1/storage/pipelines/${pipelineID}/files`;
```

### PLATFORM_URL

**Format:** Full URL (e.g., `https://platform.example.com`)

The platform's base URL. Use this to construct API endpoints.

```javascript
const platformURL = process.env.PLATFORM_URL;
// Example: "https://platform.example.com"

// Construct API URLs
const storageAPI = `${platformURL}/api/v1/storage`;
const uploadAPI = `${storageAPI}/pipelines/${pipelineID}/files`;
```

### PLATFORM_API_TOKEN

**Format:** `chrono_token_...` (pipeline-scoped token)

A pipeline-scoped API token for authenticating storage API calls. This token is automatically generated when object storage is first enabled and has limited scope (can only access storage endpoints).

```javascript
const apiToken = process.env.PLATFORM_API_TOKEN;
// Example: "chrono_token_abc123xyz789"

// Use in Authorization header
fetch(uploadURL, {
  headers: {
    'Authorization': `Bearer ${apiToken}`
  }
});
```

## How Variables Are Injected

These variables are stored in MongoDB under `config.backendEnvVars` in your pipeline document and injected into your container via Kubernetes secrets. No manual configuration is required—just enable object storage for your pipeline.

## Security Notes

- **PLATFORM_API_TOKEN** is pipeline-scoped and can only access storage endpoints
- Tokens are automatically rotated by the platform
- Never log or expose these environment variables in client-side code
- The API token should only be used server-side, never in the browser
