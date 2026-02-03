# Vue File Upload Composable

## useStorage Composable

```javascript
// composables/useStorage.js
import { ref } from 'vue';

export function useStorage(pipelineId, getToken) {
  const uploading = ref(false);
  const error = ref(null);

  const uploadFile = async (file) => {
    uploading.value = true;
    error.value = null;

    const formData = new FormData();
    formData.append('file', file);

    try {
      const response = await fetch(
        `/api/v1/storage/pipelines/${pipelineId}/files`,
        {
          method: 'POST',
          headers: { 'Authorization': `Bearer ${getToken()}` },
          body: formData,
        }
      );

      if (!response.ok) throw new Error('Upload failed');

      const data = await response.json();
      return data.url;
    } catch (err) {
      error.value = err.message;
      throw err;
    } finally {
      uploading.value = false;
    }
  };

  return { uploadFile, uploading, error };
}
```

## Usage Example

```vue
<template>
  <div>
    <input 
      type="file" 
      @change="handleFileChange" 
      :disabled="uploading" 
    />
    <p v-if="uploading">Uploading...</p>
    <p v-if="error" class="error">{{ error }}</p>
    <img v-if="uploadedUrl" :src="uploadedUrl" alt="Uploaded" />
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useStorage } from '@/composables/useStorage';

const { uploadFile, uploading, error } = useStorage('pipeline-id', () => getToken());
const uploadedUrl = ref(null);

const handleFileChange = async (e) => {
  const file = e.target.files[0];
  if (file) {
    try {
      uploadedUrl.value = await uploadFile(file);
    } catch (err) {
      console.error('Upload failed:', err);
    }
  }
};
</script>
```

## Options API Version

```javascript
// composables/useStorage.js
import { ref } from 'vue';

export function useStorage(pipelineId, getToken) {
  const uploading = ref(false);
  const error = ref(null);

  const uploadFile = async (file) => {
    uploading.value = true;
    error.value = null;

    const formData = new FormData();
    formData.append('file', file);

    try {
      const response = await fetch(
        `/api/v1/storage/pipelines/${pipelineId}/files`,
        {
          method: 'POST',
          headers: { 'Authorization': `Bearer ${getToken()}` },
          body: formData,
        }
      );

      if (!response.ok) throw new Error('Upload failed');

      const data = await response.json();
      return data.url;
    } catch (err) {
      error.value = err.message;
      throw err;
    } finally {
      uploading.value = false;
    }
  };

  return { uploadFile, uploading, error };
}
```

```vue
<template>
  <div>
    <input type="file" @change="handleFileChange" :disabled="uploading" />
  </div>
</template>

<script>
import { useStorage } from '@/composables/useStorage';

export default {
  setup() {
    const { uploadFile, uploading } = useStorage('pipeline-id', () => getToken());

    const handleFileChange = async (e) => {
      const file = e.target.files[0];
      if (file) {
        const url = await uploadFile(file);
        console.log('Uploaded:', url);
      }
    };

    return { handleFileChange, uploading };
  }
};
</script>
```
