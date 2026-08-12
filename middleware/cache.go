// middleware/cache.go
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// CacheControl sets appropriate Cache-Control headers based on the route.
//
// Rules:
//   - Static files (/public/*): aggressive caching, 1 year, immutable
//   - Homepage, category, tag, search, author: browser 1 min, CDN 1 day
//   - Post/page: browser 1 hour, CDN 7 days
//   - REST API (/api/*) and dashboard: no-store (never cached)
//   - Everything else: no-cache (fallthrough)
func CacheControl(c *fiber.Ctx) error {
	path := c.Path()

	// Static files: aggressive caching with immutable
	if strings.HasPrefix(path, "/public/") {
		c.Set("Cache-Control", "public, max-age=31536000, immutable")
		return c.Next()
	}

	// REST API and dashboard: never cache
	if strings.HasPrefix(path, "/api/") ||
		strings.HasPrefix(path, "/dashboard") ||
		strings.HasPrefix(path, "/posts") ||
		strings.HasPrefix(path, "/categories") ||
		strings.HasPrefix(path, "/tags") ||
		strings.HasPrefix(path, "/settings") ||
		strings.HasPrefix(path, "/users") ||
		strings.HasPrefix(path, "/login-") ||
		strings.HasPrefix(path, "/setup") ||
		strings.HasPrefix(path, "/logout") ||
		strings.HasPrefix(path, "/seo/") {
		c.Set("Cache-Control", "no-store")
		return c.Next()
	}

	// Sitemap & RSS feed: browser 1 hour, CDN 1 day
	if path == "/sitemap.xml" || path == "/feed.xml" || path == "/rss.xml" {
		c.Set("Cache-Control", "public, max-age=3600, s-maxage=86400")
		return c.Next()
	}

	// Dynamic manifest: short browser cache, CDN 1 day
	if path == "/manifest.json" {
		c.Set("Cache-Control", "public, max-age=60, s-maxage=86400")
		return c.Next()
	}

	// Post and page routes: browser 1 hour, CDN 7 days
	if strings.HasPrefix(path, "/post/") || strings.HasPrefix(path, "/page/") {
		c.Set("Cache-Control", "public, max-age=3600, s-maxage=604800")
		return c.Next()
	}

	// Frequently changing routes: homepage, category, tag, search, author
	// Browser 1 minute, CDN 1 day
	if path == "/" ||
		strings.HasPrefix(path, "/category/") ||
		strings.HasPrefix(path, "/tag/") ||
		strings.HasPrefix(path, "/search") ||
		strings.HasPrefix(path, "/author/") {
		c.Set("Cache-Control", "public, max-age=60, s-maxage=86400")
		return c.Next()
	}

	// Fallthrough: no-cache
	c.Set("Cache-Control", "no-cache")
	return c.Next()
}