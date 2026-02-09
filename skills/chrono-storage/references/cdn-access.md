# CDN Access

Access uploaded files through the ChronoAI platform's integrated CDN.

## URL Structure

All uploaded files are accessible via the CDN using this pattern:

```
{CHRONO_CDN_URL}/{filepath}
```

The `filepath` is returned as the `filename` field in the upload API response.

### Example

```javascript
// Upload response
const { url, filename } = await uploadFile(file);
// filename: "uploads/photos/vacation.jpg"
// url: "https://cdn.chronoai.com/my-app/uploads/photos/vacation.jpg"

// Construct URL manually
const manualUrl = `${process.env.CHRONO_CDN_URL}/${filename}`;
```

## Direct Linking

### HTML Images

```html
<!-- Direct link -->
<img src="https://cdn.chronoai.com/my-app/uploads/photo.jpg" alt="Photo">

<!-- Using environment variable in templates -->
<img src="{{ CHRONO_CDN_URL }}/uploads/photo.jpg" alt="Photo">
```

### CSS Backgrounds

```css
.hero {
  background-image: url('https://cdn.chronoai.com/my-app/assets/hero-bg.jpg');
}
```

### Video Files

```html
<video controls>
  <source src="https://cdn.chronoai.com/my-app/videos/intro.mp4" type="video/mp4">
</video>
```

### PDF Documents

```html
<a href="https://cdn.chronoai.com/my-app/documents/brochure.pdf" download>
  Download Brochure
</a>
```

## CDN Caching Behavior

The CDN automatically caches files with different TTLs based on content type:

| Content Type | Cache Duration |
|--------------|----------------|
| Images (jpg, png, gif, webp, svg) | 1 year |
| Videos (mp4, webm) | 1 year |
| Fonts (woff, woff2) | 1 year |
| CSS/JS | 1 year |
| PDFs | 1 year |
| Other files | 1 hour |

### Cache Invalidation

To invalidate cached files, upload a new file with the same path. The CDN will serve the new version within a few minutes.

For immediate invalidation, append a version query parameter:

```html
<img src="{{ fileUrl }}?v={{ timestamp }}" alt="Photo">
```

## Image Optimization

The CDN supports on-the-fly image optimization through URL parameters:

### Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `w` | Width in pixels | `?w=800` |
| `h` | Height in pixels | `?h=600` |
| `q` | Quality (1-100) | `?q=85` |
| `format` | Output format (webp, avif, jpg, png) | `?format=webp` |
| `fit` | Fit mode (cover, contain, fill) | `?fit=cover` |

### Examples

```html
<!-- Resize to 800px width -->
<img src="https://cdn.chronoai.com/my-app/uploads/photo.jpg?w=800">

<!-- Resize and convert to WebP -->
<img src="https://cdn.chronoai.com/my-app/uploads/photo.jpg?w=800&format=webp">

<!-- High quality thumbnail -->
<img src="https://cdn.chronoai.com/my-app/uploads/photo.jpg?w=200&h=200&q=90&fit=cover">

<!-- Responsive image with WebP fallback -->
<picture>
  <source srcset="photo.jpg?format=webp&w=800" type="image/webp">
  <img src="https://cdn.chronoai.com/my-app/uploads/photo.jpg?w=800" alt="Photo">
</picture>
```

## Browser Access Examples

### React

```jsx
function ImageDisplay({ filename, alt }) {
  return (
    <img src={`${process.env.CHRONO_CDN_URL}/${filename}`} alt={alt} />
  );
}

// With optimization
function OptimizedImage({ filename, alt, width }) {
  const src = `${process.env.CHRONO_CDN_URL}/${filename}?w=${width}&format=webp`;
  return <img src={src} alt={alt} />;
}
```

### Vue.js

```vue
<template>
  <img :src="fullUrl" :alt="alt">
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps(['filename', 'alt']);

const fullUrl = computed(() =>
  `${import.meta.env.CHRONO_CDN_URL}/${props.filename}`
);
</script>
```

### Next.js Image Component

```jsx
import Image from 'next/image';

export function StorageImage({ filename, alt, width, height }) {
  const src = `${process.env.CHRONO_CDN_URL}/${filename}`;

  return (
    <Image
      src={src}
      alt={alt}
      width={width}
      height={height}
      unoptimized // CDN handles optimization
    />
  );
}
```

## Tips

1. **Use HTTPS** - Always use HTTPS URLs for CDN resources
2. **Lazy load images** - Add `loading="lazy"` for below-fold images
3. **Preload critical assets** - Use `<link rel="preload">` for above-fold images
4. **Pick the right format** - Use WebP for photos, PNG for graphics with transparency
5. **Profile your images** - Use browser dev tools to find optimal quality settings
