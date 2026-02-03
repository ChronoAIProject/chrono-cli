# Class-Based ChronoStorage Utility

## ChronoStorage Class

```javascript
class ChronoStorage {
  constructor(pipelineId, getToken) {
    this.pipelineId = pipelineId;
    this.getToken = getToken;
  }

  async upload(file) {
    const formData = new FormData();
    formData.append('file', file);

    const response = await fetch(
      `/api/v1/storage/pipelines/${this.pipelineId}/files`,
      {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.getToken()}`,
        },
        body: formData,
      }
    );

    if (!response.ok) {
      throw new Error(`Upload failed: ${response.statusText}`);
    }

    const data = await response.json();
    return data.url; // CDN URL
  }

  async list() {
    const response = await fetch(
      `/api/v1/storage/pipelines/${this.pipelineId}/files`,
      {
        headers: {
          'Authorization': `Bearer ${this.getToken()}`,
        },
      }
    );

    if (!response.ok) {
      throw new Error(`List failed: ${response.statusText}`);
    }

    return await response.json();
    // Returns: { files: ["photo.jpg", "doc.pdf"], total: 2 }
  }

  async delete(filename) {
    const response = await fetch(
      `/api/v1/storage/pipelines/${this.pipelineId}/files/${filename}`,
      {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${this.getToken()}`,
        },
      }
    );

    if (!response.ok) {
      throw new Error(`Delete failed: ${response.statusText}`);
    }

    return await response.json();
    // Returns: { message: "File deleted successfully", filename: "photo.jpg" }
  }

  getCDNUrl(filename) {
    // CHRONO_CDN_URL is injected as: https://cdn.chrono-ai.fun/{appName}
    const baseUrl = process.env.CHRONO_CDN_URL;
    return `${baseUrl}/${filename}`;
  }
}
```

## Usage Examples

### Basic Upload

```javascript
const storage = new ChronoStorage('pipeline-id', () => getToken());

try {
  const cdnUrl = await storage.upload(file);
  console.log('File available at:', cdnUrl);
} catch (error) {
  console.error('Upload failed:', error);
}
```

### Upload with CDN URL Construction

```javascript
const storage = new ChronoStorage('pipeline-id', () => getToken());

async function uploadAndDisplay(file) {
  const filename = file.name;
  const cdnUrl = await storage.upload(file);
  
  // Alternative: construct CDN URL manually
  // const manualUrl = storage.getCDNUrl(filename);
  
  return { filename, cdnUrl };
}
```

### List and Delete Files

```javascript
const storage = new ChronoStorage('pipeline-id', () => getToken());

// List all files
const { files, total } = await storage.list();
console.log(`Found ${total} files:`, files);

// Delete a specific file
await storage.delete('photo.jpg');
```

### TypeScript Version

```typescript
interface UploadResponse {
  filename: string;
  url: string;
  size: number;
}

interface ListResponse {
  files: string[];
  total: number;
}

interface DeleteResponse {
  message: string;
  filename: string;
}

type AuthTokenGetter = () => string;

class ChronoStorage {
  constructor(
    private pipelineId: string,
    private getToken: AuthTokenGetter
  ) {}

  async upload(file: File): Promise<string> {
    const formData = new FormData();
    formData.append('file', file);

    const response = await fetch(
      `/api/v1/storage/pipelines/${this.pipelineId}/files`,
      {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.getToken()}`,
        },
        body: formData,
      }
    );

    if (!response.ok) {
      throw new Error(`Upload failed: ${response.statusText}`);
    }

    const data: UploadResponse = await response.json();
    return data.url;
  }

  async list(): Promise<ListResponse> {
    const response = await fetch(
      `/api/v1/storage/pipelines/${this.pipelineId}/files`,
      {
        headers: {
          'Authorization': `Bearer ${this.getToken()}`,
        },
      }
    );

    if (!response.ok) {
      throw new Error(`List failed: ${response.statusText}`);
    }

    return await response.json();
  }

  async delete(filename: string): Promise<DeleteResponse> {
    const response = await fetch(
      `/api/v1/storage/pipelines/${this.pipelineId}/files/${filename}`,
      {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${this.getToken()}`,
        },
      }
    );

    if (!response.ok) {
      throw new Error(`Delete failed: ${response.statusText}`);
    }

    return await response.json();
  }

  getCDNUrl(filename: string): string {
    const baseUrl = process.env.CHRONO_CDN_URL || '';
    return `${baseUrl}/${filename}`;
  }
}
```
