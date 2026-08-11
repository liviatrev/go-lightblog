// utils/cloudflare.go
package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go-lightblog/database"
	"go-lightblog/models"
)

// ============================================================
// CLOUDFLARE CACHE PURGE HELPERS
// ============================================================

// IsCloudflareEnabled checks if Cloudflare purge is enabled in settings
// and both API key and Zone ID are configured.
func IsCloudflareEnabled() bool {
	enableCloudflare := models.GetSetting(database.DB, "enable_cloudflare", "no")
	apiKey := models.GetSetting(database.DB, "cloudflare_api_key", "")
	zoneID := models.GetSetting(database.DB, "cloudflare_zone_id", "")

	return enableCloudflare == "yes" && apiKey != "" && zoneID != ""
}

// GetSiteURL returns the public site URL from settings.
// Returns empty string if not configured.
func GetSiteURL() string {
	return models.GetSetting(database.DB, "site_url", "")
}

// LogCloudflareError logs a Cloudflare purge error with context.
// Used in goroutines so purge failures don't block the main response path.
func LogCloudflareError(context string, err error) {
	if err != nil {
		log.Printf("Cloudflare purge error (%s): %v", context, err)
	}
}

// purgeRequest is the payload sent to the Cloudflare API for purge by URL.
type purgeRequest struct {
	Files []string `json:"files"`
}

// purgeCloudflareURLs purges a list of URLs from the Cloudflare cache.
// It uses the "Purge by URL" endpoint (not purge everything),
// allowing a single request to purge multiple URLs (post + categories + tags).
func purgeCloudflareURLs(urls []string) error {
	if !IsCloudflareEnabled() {
		return nil
	}

	apiKey := models.GetSetting(database.DB, "cloudflare_api_key", "")
	zoneID := models.GetSetting(database.DB, "cloudflare_zone_id", "")

	// Filter out empty URLs
	var cleanURLs []string
	for _, u := range urls {
		if u != "" {
			cleanURLs = append(cleanURLs, u)
		}
	}
	if len(cleanURLs) == 0 {
		return nil
	}

	body, err := json.Marshal(purgeRequest{Files: cleanURLs})
	if err != nil {
		return fmt.Errorf("failed to marshal purge request: %v", err)
	}

	endpoint := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/purge_cache", zoneID)

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create purge request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send purge request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("cloudflare API returned status %d", resp.StatusCode)
	}

	log.Printf("Cloudflare cache purged for %d URL(s): %v", len(cleanURLs), cleanURLs)
	return nil
}

// buildPostPurgeURLs constructs the list of URLs to purge for a given post.
// It loads the post's category and tags from the database if not preloaded.
func buildPostPurgeURLs(post models.Post) []string {
	baseURL := GetSiteURL()
	if baseURL == "" || !IsCloudflareEnabled() {
		return nil
	}

	// Reload the post with relationships if the slug is empty
	if post.ID > 0 && post.Category.Slug == "" && len(post.Tags) == 0 {
		var fullPost models.Post
		if err := database.DB.Preload("Category").Preload("Tags").First(&fullPost, post.ID).Error; err == nil {
			post = fullPost
		}
	}

	// Normalize base URL (strip trailing slash)
	for len(baseURL) > 1 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	urls := []string{
		baseURL + "/",
	}

	// Post/page URL
	if post.Type == "page" {
		urls = append(urls, baseURL+"/page/"+post.Slug)
	} else {
		urls = append(urls, baseURL+"/post/"+post.Slug)
	}

	// Category URL
	if post.Category.Slug != "" {
		urls = append(urls, baseURL+"/category/"+post.Category.Slug)
	}

	// Tag URLs
	for _, tag := range post.Tags {
		if tag.Slug != "" {
			urls = append(urls, baseURL+"/tag/"+tag.Slug)
		}
	}

	return urls
}

// PurgePostCache purges cache for a post (by ID) and its associated
// category, tags, and homepage. It loads the post's relationships
// from the database to build the correct purge URL list.
func PurgePostCache(postID uint) error {
	var post models.Post
	if err := database.DB.Preload("Category").Preload("Tags").First(&post, postID).Error; err != nil {
		return err
	}

	urls := buildPostPurgeURLs(post)
	return purgeCloudflareURLs(urls)
}

// PurgePostCacheByPost purges cache for a post using an already-loaded post object.
// This avoids re-querying the database and works with soft-deleted posts.
func PurgePostCacheByPost(post models.Post) error {
	urls := buildPostPurgeURLs(post)
	return purgeCloudflareURLs(urls)
}

// PurgeSitemapCache purges the sitemap.xml URL from Cloudflare cache.
func PurgeSitemapCache() error {
	baseURL := GetSiteURL()
	if baseURL == "" || !IsCloudflareEnabled() {
		return nil
	}

	for len(baseURL) > 1 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	return purgeCloudflareURLs([]string{baseURL + "/sitemap.xml"})
}

// PurgePostCacheBySlug purges cache for a post using its slug.
// Useful when the post object isn't fully loaded in the handler.
func PurgePostCacheBySlug(slug, postType string) error {
	var post models.Post
	if err := database.DB.Where("slug = ?", slug).Preload("Category").Preload("Tags").First(&post).Error; err != nil {
		return err
	}
	if postType != "" {
		post.Type = postType
	}

	urls := buildPostPurgeURLs(post)
	return purgeCloudflareURLs(urls)
}

// PurgeTaxonomyCache purges cache for a category or tag page and the homepage.
// kind can be "category" or "tag".
func PurgeTaxonomyCache(slug, kind string) error {
	baseURL := GetSiteURL()
	if baseURL == "" || !IsCloudflareEnabled() {
		return nil
	}

	for len(baseURL) > 1 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	urls := []string{
		baseURL + "/",
		baseURL + "/" + kind + "/" + slug,
	}

	return purgeCloudflareURLs(urls)
}