# Error Handling

Common errors when using the chrono-storage API and how to handle them.

## HTTP Status Codes

| Status | Error Name | Description |
|--------|------------|-------------|
| 200 | OK | Upload successful |
| 400 | Bad Request | Invalid request format |
| 401 | Unauthorized | Invalid or expired token |
| 403 | Forbidden | Token lacks storage scope |
| 413 | Payload Too Large | File exceeds size limit |
| 415 | Unsupported Media Type | Invalid content type |
| 500 | Server Error | Platform error (retry) |

## Common Errors

### 401 Unauthorized

**Cause:** The API token is invalid, expired, or malformed.

**Response:**
```json
{
  "error": "Unauthorized",
  "message": "Invalid or expired API token"
}
```

**Solutions:**
1. Verify `PLATFORM_API_TOKEN` is set correctly in environment variables
2. Check token hasn't expired (tokens have a 1-year lifetime)
3. Ensure the token format is correct: `chr_sk_live_...`
4. Regenerate token through the platform dashboard if needed

**Example Handler:**
```javascript
if (response.status === 401) {
  console.error('API token is invalid or expired');
  // Prompt user to reconfigure or contact admin
}
```

---

### 403 Forbidden

**Cause:** The token exists but doesn't have storage permissions.

**Response:**
```json
{
  "error": "Forbidden",
  "message": "Token lacks storage:write scope"
}
```

**Solutions:**
1. Verify the token was created with storage permissions
2. Check that the token belongs to the correct pipeline
3. Regenerate token with proper scopes through the dashboard

**Example Handler:**
```javascript
if (response.status === 403) {
  console.error('Token does not have storage permissions');
  // Contact platform administrator
}
```

---

### 413 Payload Too Large

**Cause:** The uploaded file exceeds the size limit (default 50MB).

**Response:**
```json
{
  "error": "Payload Too Large",
  "message": "File size exceeds maximum allowed size of 52428800 bytes"
}
```

**Solutions:**
1. Validate file size on the client before uploading
2. Compress or optimize files before upload
3. For larger files, contact support to increase limits (max 100MB)

**Client-Side Validation:**
```javascript
const MAX_SIZE = 50 * 1024 * 1024; // 50MB

function validateFileSize(file) {
  if (file.size > MAX_SIZE) {
    alert(`File is too large. Maximum size is 50MB.`);
    return false;
  }
  return true;
}

// Usage
fileInput.addEventListener('change', (e) => {
  const file = e.target.files[0];
  if (file && validateFileSize(file)) {
    uploadFile(file);
  }
});
```

---

### 415 Unsupported Media Type

**Cause:** The file type is not allowed or not properly detected.

**Response:**
```json
{
  "error": "Unsupported Media Type",
  "message": "Content type 'application/x-msdownload' is not allowed"
}
```

**Solutions:**
1. Verify the file has a valid MIME type
2. Check that the file extension matches the content
3. Ensure you're uploading allowed file types

**Client-Side Validation:**
```javascript
const ALLOWED_TYPES = [
  'image/jpeg',
  'image/png',
  'image/gif',
  'image/webp',
  'application/pdf',
  'video/mp4'
];

function validateFileType(file) {
  if (!ALLOWED_TYPES.includes(file.type)) {
    alert(`File type ${file.type} is not allowed.`);
    return false;
  }
  return true;
}
```

---

### 500 Server Error

**Cause:** Temporary platform issue or internal error.

**Response:**
```json
{
  "error": "Internal Server Error",
  "message": "An unexpected error occurred"
}
```

**Solutions:**
1. Implement retry logic with exponential backoff
2. Check platform status page for ongoing issues
3. Report persistent errors to platform support

**Retry Handler:**
```javascript
async function uploadWithRetry(file, maxRetries = 3) {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      return await uploadFile(file);
    } catch (error) {
      if (error.response?.status === 500 && attempt < maxRetries - 1) {
        const delay = Math.pow(2, attempt) * 1000; // 1s, 2s, 4s
        await new Promise(resolve => setTimeout(resolve, delay));
        continue;
      }
      throw error;
    }
  }
}
```

---

## Error Response Format

All error responses follow this structure:

```json
{
  "error": "Error Name",
  "message": "Detailed error message",
  "requestId": "abc-123-def-456"
}
```

The `requestId` is useful for debugging and support inquiries.

---

## Best Practices

### 1. Always Handle Errors

```javascript
try {
  const result = await uploadFile(file);
  return result;
} catch (error) {
  // Log error details
  console.error('Upload failed:', {
    status: error.response?.status,
    message: error.response?.data?.message,
    requestId: error.response?.data?.requestId
  });

  // Return user-friendly message
  throw new Error('Failed to upload file. Please try again.');
}
```

### 2. Provide User Feedback

```javascript
async function handleUpload(file) {
  showLoadingIndicator();

  try {
    const result = await uploadWithRetry(file);
    showSuccessMessage('File uploaded successfully!');
    return result;
  } catch (error) {
    let message = 'Upload failed. ';

    switch (error.response?.status) {
      case 401:
        message += 'Please re-authenticate.';
        break;
      case 413:
        message += 'File is too large.';
        break;
      case 415:
        message += 'File type not supported.';
        break;
      default:
        message += 'Please try again.';
    }

    showErrorMessage(message);
  } finally {
    hideLoadingIndicator();
  }
}
```

### 3. Log for Debugging

```javascript
function logUploadError(error, file) {
  const logData = {
    timestamp: new Date().toISOString(),
    fileName: file.name,
    fileSize: file.size,
    fileType: file.type,
    status: error.response?.status,
    message: error.response?.data?.message,
    requestId: error.response?.data?.requestId
  };

  // Send to logging service
  logger.error('Storage upload failed', logData);
}
```

### 4. Graceful Degradation

```javascript
async function uploadOrFallback(file) {
  try {
    return await uploadFile(file);
  } catch (error) {
    // Fallback to local storage or alternative service
    console.warn('Platform upload failed, using fallback:', error.message);
    return await fallbackUploadService(file);
  }
}
```
