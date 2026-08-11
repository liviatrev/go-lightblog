# go-lightblog

`go-lightblog` is a high-performance, lightweight, and modern blogging engine and headless CMS built with **Go (Golang)** and the **Fiber v2** web framework. It uses **SQLite** as its database with **GORM** (CGO-free driver), integrating modern development practices like automated Gemini AI-powered SEO generation, a custom asset cache-busting system, WebP/JPEG image processing, and Cloudflare CDN caching/purging.

Designed to run either as a traditional web-rendered blog or as a headless CMS with a fully featured JSON REST API, it can be hosted easily over TCP or Unix Sockets.

---

## 🚀 Key Features

*   **⚡ Ultra-Fast Performance**: Built on Fiber v2 (Go's fastest HTTP engine) with optimized connection pooling.
*   **📂 Zero-CGO SQLite Database**: Database powered by pure-Go SQLite with GORM, allowing easy setup and distribution without external C dependencies.
*   **🔒 Secure Setup & Obfuscated Login**:
    *   One-time `/setup` wizard to bootstrap the blog and database.
    *   Unique obfuscated login path `/login-:token` protecting against brute-force and discovery attacks.
*   **👥 Role-Based Access Control (RBAC)**: Two separate privilege levels:
    *   `admin`: Full access to settings, user management, and API controls.
    *   `editor`: Access to posts, category, and tag management.
*   **🤖 Automated AI SEO**: Integrated Google Gemini AI to analyze article body content and automatically generate `MetaTitle`, `MetaDescription`, and `TargetKeyword`.
*   **🖼️ Intelligent Media Processing Pipeline**:
    *   Supports dynamic, on-demand image resizing with JPG & lossy WebP formats using zero-CGO tools.
    *   Optional seamless integration with **ImageKit CDN** for media asset hosting.
    *   Supports `<picture>` tags with WebP-first fallback mechanisms.
*   **🌀 High-Performance Cache Buster**: Dynamic in-memory MD5-hash cache-busting (`?v=abc12345`) for CSS/JS static files with standard RWMutex.
*   **🌐 Aggressive Cache Control & Cloudflare Purging**:
    *   Fine-grained `Cache-Control` header middleware customized for static assets, dynamic templates, pages, and API endpoints.
    *   Async, URL-specific Cloudflare CDN cache purging triggered upon creating or modifying content.
*   **🎨 Public Color Themes**: 6 beautiful built-in CSS themes (Light, Ocean, Forest, Sunset, Midnight Dark, and Royal Purple) based on Bootstrap 5.3 custom variables.
*   **✏️ Modern WYSIWYG Editor**: Uses the feature-rich SunEditor for editing posts in the administration panel.

---

## 🛠️ Technology Stack & Dependencies

*   **Backend Language**: Go (v1.25.0)
*   **Web Framework**: [Fiber v2](https://github.com/gofiber/fiber) (for routing & middleware)
*   **Template Engine**: [gofiber/template/html/v2](https://github.com/gofiber/template) (HTML parsing with layout support)
*   **Database & ORM**: [GORM](https://gorm.io/) with [glebarez/sqlite](https://github.com/glebarez/sqlite) (CGO-free driver)
*   **AI Engine**: [Google GenAI SDK](https://google.golang.org/genai) (Gemini models)
*   **Media Transforms**: [disintegration/imaging](https://github.com/disintegration/imaging) and [deepteams/webp](https://github.com/deepteams/webp) (lossy VP8 WebP encoder)
*   **CDN Integration**: [imagekit-go SDK v2](https://github.com/imagekit-developer/imagekit-go)
*   **Hot Reloading**: [Air](https://github.com/cosmtrek/air) for dev servers

---

## 📁 Repository Structure

```
go-lightblog/
├── config/              # App & Session store configurations
├── database/            # Database connection & migration setup
├── handlers/            # HTTP handlers (separating view controllers & JSON API endpoints)
│   ├── api_*.go         # Headless REST API controller logic (Users, Posts, Settings, etc.)
│   ├── auth.go          # Session-based authentication & logins
│   └── public.go        # Public-facing views (home, post readers, thumbnails proxy)
├── middleware/          # Security, Auth, Cache-Control, Setup-checks
├── models/              # GORM Database schemas & setting repositories
├── public/              # Static public resources (compiled JS/CSS, local uploads)
│   ├── css/themes/      # Built-in palette color themes
│   └── uploads/         # Local folder path for media assets
├── utils/               # Utilities (cachebuster, Cloudflare API wrapper, helpers)
├── views/               # Layouts and HTML views templates
│   ├── dashboard/       # Administration panel HTML files
│   └── layouts/         # Layout definitions (public, admin layouts)
├── .air.toml            # Hot reload server configurations
├── .env                 # Environment variables config
├── .gitignore           # Ignored files (temporary artifacts, .db, uploads)
├── go.mod / go.sum      # Go module dependencies
├── main.go              # CLI parameters, Template registrations & Application entry point
├── API_DOCUMENTATION.md # Detailed JSON API guides
├── CACHE_BUSTER_IMPLEMENTATION.md # Design documentation for static assets cache-busting
└── CHANGELOG.md         # Full project history & version details
```

---

## ⚙️ Setup & Configuration

### 1. Prerequisites
Make sure you have [Go](https://go.dev/) installed (version 1.25+ is recommended).

### 2. Environment Variables (`.env`)
Create a `.env` file in the root directory to store your private cloud API keys. You can configure **ImageKit** for image CDN storage:

```env
IMAGEKIT_PUBLIC_KEY=your_imagekit_public_key
IMAGEKIT_PRIVATE_KEY=your_imagekit_private_key
IMAGEKIT_URL_ENDPOINT=https://ik.imagekit.io/your_endpoint
```

*Note: If ImageKit configurations are omitted or set to local mode, `go-lightblog` will save images locally in `./public/uploads` and resize/serve them locally.*

### 3. CLI Run Parameters
The binary accepts multiple parameters for port binds, UNIX socket modes, and database files:

| Flag | Default | Description |
| :--- | :--- | :--- |
| `-l` | `""` | Listen TCP address (e.g., `127.0.0.1` or `0.0.0.0`) |
| `-p` | `5800` | Listen TCP port (used if `-l` or `-p` is provided) |
| `-sock` | `/tmp/lightblog.sock` | Path to Unix socket file (used if no TCP parameters are provided) |
| `-db` | `lightblog.db` | Path to the SQLite database file |
| `-a` | `./public` | Public folder path (where static and local assets reside) |

### 4. Running the Application

**Development Mode (Hot Reload with Air):**
```bash
air
```

**Standard Run:**
```bash
go run main.go -p 5800 -db lightblog.db
```

---

## 🌀 Key Architectural Workflows

### 🛠️ 1. First-Time Setup Wizard
1. When you run `go-lightblog` for the first time, accessing any URL will trigger the `CheckSetup` middleware and redirect you to `/setup`.
2. Enter your site name, desired admin credentials (username & password), and choose a unique, secure **Login Token**.
3. Upon completion, the setup wizard creates the initial administrative user and locks the `/setup` route.

### 🔑 2. Login URL Obfuscation
Rather than having a standard `/admin` or `/login` path which is vulnerable to malicious bot crawlers, `go-lightblog` uses a dynamic path configured during setup:
*   **Login Page URL**: `/login-<your_secret_token>`
*   **Login Action POST**: `/login-<your_secret_token>/process`

If the correct token is not supplied in the URL, the system returns a standard `404 Not Found`.

### 👥 3. Headless REST API vs HTML Admin Panel
`go-lightblog` serves dual purposes:
1. **Server-Side Rendered (SSR) Frontend**: The app renders templates inside `views/` using the custom-built UI, session cookie-based logins, and HTML.
2. **Headless Admin REST API**: Every admin capability has a parallel JSON endpoint mapped under `/api/v1/admin/*`.
   *   Authenticates via a secure `API Key` (Bearer token) generated for each user.
   *   Supports RBAC (e.g. users with the `editor` role are forbidden from accessing `/users` or `/settings`).
   *   Perfect for developers wanting to build Jamstack frontends (Next.js, Nuxt, Astro) or mobile client applications.

*See `API_DOCUMENTATION.md` for exact paths, parameter schemas, and payload samples.*

### 🖼️ 4. Local WebP Resizing & CDN Thumbnail Pipeline
To achieve extreme Lighthouse scores, images are compressed and resized automatically:
*   **ImageKit CDN**: If active, templates request URLs appended with real-time ImageKit optimization queries (e.g. `?tr=w-600,q-70,f-webp` and `f-jpg`).
*   **Local Image Proxy**: If hosted locally, images are served through the proxy `/api/thumb?src=<source>&w=<width>&f=<webp|jpg>`.
    *   This proxy uses `github.com/deepteams/webp` and `disintegration/imaging` to resize pictures on-the-fly and convert transparent elements to solid background colors, reducing bandwidth usage.
    *   Calculated results are saved to disk, ensuring future lookups are high-speed hits with aggressive cache-control headers (`max-age=31536000`).

### 🌐 5. Caching & Cloudflare Cache Purging
To maximize CDN performance while retaining up-to-date dynamics:
*   **Cache Middleware**: Static files `/public/*` are served with `Cache-Control: public, max-age=31536000, immutable`. Individual blog posts `/post/:slug` are cached on the browser for 1 hour and CDNs for 7 days.
*   **Asynchronous Purging**: If Cloudflare settings are configured, any publication or update of categories, tags, or articles triggers an **async, non-blocking HTTP purge request** containing precise URLs (e.g., the specific post URL, category URL, and homepage). It only clears specific modified URLs, leaving the rest of your CDN cached assets untouched.

### 🎨 6. Dynamic Theme Palette Switcher
Admin users can change the public blog theme globally inside the CMS. Themes are built with CSS customized variables:
*   `light` (Default light design)
*   `ocean` (Calm blue theme)
*   `forest` (Earth organic green)
*   `sunset` (Warm orange/yellow palettes)
*   `midnight` (Contrast dark theme)
*   `royal` (Rich dark purple accents)

The layouts dynamically inject the active stylesheet at runtime through the custom `themeURL` HTML template function.

---

## 🧪 Testing and Quality
You can run automated tests using Go's built-in tool:
```bash
go test ./...
```

---

## 📜 License & Acknowledgments
This project is open-source. For information regarding changes and feature histories, please check out the `CHANGELOG.md` file. Detailed guides for REST API endpoints are located in `API_DOCUMENTATION.md`.
