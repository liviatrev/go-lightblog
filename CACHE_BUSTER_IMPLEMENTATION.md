# Cache Buster Implementation for go-lightblog

## Summary
Implemented a cache buster mechanism for static files (.css, .js) to prevent browser caching issues when files are updated.

## Files Created
1. **utils/cachebuster.go** - Core cache buster utility with:
   - `GetCacheBuster(filePath string) string` - Returns MD5 hash (first 8 chars) of file content
   - `CacheBusterURL(filePath string) string` - Returns URL with cache buster query parameter (e.g., `/public/css/admin.min.css?v=abc12345`)
   - Thread-safe caching with `sync.RWMutex`
   - Path traversal protection

## Files Modified
1. **main.go** - Added import for utils package and registered template functions:
   - `cacheBuster` - Returns just the hash
   - `cacheBusterURL` - Returns full URL with cache buster

2. **views/layouts/main.html** - Admin layout:
   - CSS: `<link href="{{cacheBusterURL "/public/css/admin.min.css"}}" rel="stylesheet">`
   - JS: `<script src="{{cacheBusterURL "/public/js/admin.min.js"}}"></script>`

3. **views/layouts/public.html** - Public layout:
   - CSS: `<link href="{{cacheBusterURL "/public/css/public.min.css"}}" rel="stylesheet">`

4. **views/login.html** - Login page:
   - CSS: `<link href="{{cacheBusterURL "/public/css/auth.min.css"}}" rel="stylesheet">`

5. **views/setup.html** - Setup page:
   - CSS: `<link href="{{cacheBusterURL "/public/css/auth.min.css"}}" rel="stylesheet">`

6. **views/dashboard/posts_create.html** - Post create form:
   - CSS: `<link href="{{cacheBusterURL "/public/css/posts-form.min.css"}}" rel="stylesheet">`
   - JS: `<script src="{{cacheBusterURL "/public/js/posts-form.min.js"}}"></script>`

7. **views/dashboard/posts_edit.html** - Post edit form:
   - CSS: `<link href="{{cacheBusterURL "/public/css/posts-form.min.css"}}" rel="stylesheet">`
   - JS: `<script src="{{cacheBusterURL "/public/js/posts-form.min.js"}}"></script>`

8. **views/dashboard/settings.html** - Settings page:
   - JS: `<script src="{{cacheBusterURL "/public/js/settings.min.js"}}"></script>`

9. **views/post.html** - Post view page:
   - JS: `<script src="{{cacheBusterURL "/public/js/post.min.js"}}"></script>`

## How It Works
- When a template is rendered, `cacheBusterURL` computes an MD5 hash of the file content
- The hash is cached in memory to avoid recomputing on every request
- The URL becomes: `/public/css/admin.min.css?v=a1b2c3d4`
- When the file content changes, the hash changes, forcing browsers to fetch the new version
- If file doesn't exist, returns original URL without cache buster

## Usage in Templates
```html
<!-- For CSS -->
<link href="{{cacheBusterURL "/public/css/your-file.min.css"}}" rel="stylesheet">

<!-- For JS -->
<script src="{{cacheBusterURL "/public/js/your-file.min.js"}}"></script>

<!-- Or just get the hash -->
