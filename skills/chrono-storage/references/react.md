# React File Upload Hook

## useFileUpload Hook

```javascript
import { useState } from 'react';

function useFileUpload(pipelineId, getToken) {
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);

  const uploadFile = async (file) => {
    setUploading(true);

    const formData = new FormData();
    formData.append('file', file);

    try {
      const response = await fetch(
        `/api/v1/storage/pipelines/${pipelineId}/files`,
        {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${getToken()}`,
          },
          body: formData,
        }
      );

      if (!response.ok) throw new Error('Upload failed');

      const { url } = await response.json();
      return url;
    } catch (error) {
      console.error('Upload error:', error);
      throw error;
    } finally {
      setUploading(false);
    }
  };

  return { uploadFile, uploading };
}
```

## Usage Example

```javascript
function UploadComponent() {
  const { uploadFile, uploading } = useFileUpload('pipeline-id', () => getToken());

  const handleFileChange = async (e) => {
    const file = e.target.files[0];
    if (file) {
      const cdnUrl = await uploadFile(file);
      console.log('File uploaded:', cdnUrl);
    }
  };

  return (
    <input type="file" onChange={handleFileChange} disabled={uploading} />
  );
}
```

## With Progress Tracking

For upload progress, use XMLHttpRequest:

```javascript
function useFileUpload(pipelineId, getToken) {
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);

  const uploadFile = (file) => {
    return new Promise((resolve, reject) => {
      setUploading(true);
      setProgress(0);

      const formData = new FormData();
      formData.append('file', file);

      const xhr = new XMLHttpRequest();
      xhr.upload.addEventListener('progress', (e) => {
        if (e.lengthComputable) {
          setProgress(Math.round((e.loaded / e.total) * 100));
        }
      });

      xhr.addEventListener('load', () => {
        if (xhr.status === 201) {
          const { url } = JSON.parse(xhr.responseText);
          resolve(url);
        } else {
          reject(new Error('Upload failed'));
        }
        setUploading(false);
      });

      xhr.addEventListener('error', () => {
        reject(new Error('Network error'));
        setUploading(false);
      });

      xhr.open('POST', `/api/v1/storage/pipelines/${pipelineId}/files`);
      xhr.setRequestHeader('Authorization', `Bearer ${getToken()}`);
      xhr.send(formData);
    });
  };

  return { uploadFile, uploading, progress };
}
```
