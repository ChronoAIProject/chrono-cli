# Upload API

## Platform API Upload

The simplest way to upload files is through the platform's storage API. Your backend server handles the upload and receives the CDN URL.

## Endpoint

```
POST {PLATFORM_URL}/api/v1/storage/pipelines/{PIPELINE_ID}/files
```

## Headers

| Header          | Value                        |
|-----------------|------------------------------|
| Authorization   | `Bearer {PLATFORM_API_TOKEN}` |

## Request Body

`multipart/form-data` with a `file` field containing the file to upload.

## Response

```json
{
  "url": "https://cdn.example.com/myapp/uploads/photo.jpg",
  "filename": "photo.jpg",
  "size": 245632,
  "contentType": "image/jpeg"
}
```

## Code Examples

### Node.js (Express)

```javascript
import express from 'express';
import formData from 'form-data';
import fetch from 'node-fetch';
import { createReadStream } from 'fs';

const app = express();

app.post('/api/upload', async (req, res) => {
  const { platformURL, pipelineID, apiToken } = process.env;

  // Get file from request
  const file = req.files?.file;
  if (!file) {
    return res.status(400).json({ error: 'No file provided' });
  }

  // Upload to platform storage
  const form = new formData();
  form.append('file', createReadStream(file.path), {
    filename: file.name,
    contentType: file.mimetype
  });

  const response = await fetch(
    `${platformURL}/api/v1/storage/pipelines/${pipelineID}/files`,
    {
      method: 'POST',
      headers: {
        ...form.getHeaders(),
        'Authorization': `Bearer ${apiToken}`
      },
      body: form
    }
  );

  if (!response.ok) {
    const error = await response.json();
    return res.status(response.status).json(error);
  }

  const result = await response.json();
  res.json(result);
});
```

### Next.js (App Router)

```javascript
// app/api/upload/route.js
import { NextResponse } from 'next/server';
import formData from 'form-data';
import fetch from 'node-fetch';

export async function POST(request) {
  const { platformURL, pipelineID, apiToken } = process.env;

  const formData = await request.formData();
  const file = formData.get('file');

  if (!file) {
    return NextResponse.json(
      { error: 'No file provided' },
      { status: 400 }
    );
  }

  const form = new formData();
  form.append('file', file.stream(), file.name);

  const response = await fetch(
    `${platformURL}/api/v1/storage/pipelines/${pipelineID}/files`,
    {
      method: 'POST',
      headers: {
        ...form.getHeaders(),
        'Authorization': `Bearer ${apiToken}`
      },
      body: form
    }
  );

  const result = await response.json();
  return NextResponse.json(result);
}
```

### Vue.js + Express Backend

```javascript
// Backend (Express)
app.post('/api/upload', async (req, res) => {
  const { platformURL, pipelineID, apiToken } = process.env;
  const form = new formData();
  form.append('file', req.files.file.data, req.files.file.name);

  const response = await fetch(
    `${platformURL}/api/v1/storage/pipelines/${pipelineID}/files`,
    {
      method: 'POST',
      headers: {
        ...form.getHeaders(),
        'Authorization': `Bearer ${apiToken}`
      },
      body: form
    }
  );

  res.json(await response.json());
});

// Frontend (Vue)
async uploadFile(file) {
  const formData = new FormData();
  formData.append('file', file);

  const response = await fetch('/api/upload', {
    method: 'POST',
    body: formData
  });

  const { url } = await response.json();
  return url;
}
```

### Vanilla JS + Backend

```javascript
// Frontend
async function uploadFile(file) {
  const formData = new FormData();
  formData.append('file', file);

  const response = await fetch('/api/upload', {
    method: 'POST',
    body: formData
  });

  const { url } = await response.json();
  console.log('File uploaded:', url);
  return url;
}

// HTML
<input type="file" onchange="uploadFile(this.files[0])">
```

## File Size Limits

- Default: 50 MB per file
- Configurable per pipeline
- Returns `413 Payload Too Large` if exceeded

## Supported File Types

All file types are supported, including:
- Images: `.jpg`, `.png`, `.gif`, `.webp`, `.svg`
- Documents: `.pdf`, `.doc`, `.docx`, `.txt`
- Videos: `.mp4`, `.webm`, `.mov`
- Archives: `.zip`, `.tar`, `.gz`
- Any other binary or text file
