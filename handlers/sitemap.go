// handlers/sitemap.go
package handlers

import (
	"encoding/xml"
	"fmt"
	"time"

	"go-lightblog/database"
	"go-lightblog/models"
	"go-lightblog/utils"

	"github.com/gofiber/fiber/v2"
)

type SitemapURL struct {
	XMLName    xml.Name `xml:"url"`
	Loc        string   `xml:"loc"`
	LastMod    string   `xml:"lastmod"`
	ChangeFreq string   `xml:"changefreq"`
	Priority   float64  `xml:"priority"`
}

type SitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

// Sitemap renders a dynamic XML sitemap of the website
func Sitemap(c *fiber.Ctx) error {
	baseURL := utils.GetSiteURL()
	if baseURL == "" {
		baseURL = c.BaseURL()
	}

	// Normalize base URL (strip trailing slash)
	for len(baseURL) > 1 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	var urls []SitemapURL

	// 1. Homepage
	urls = append(urls, SitemapURL{
		Loc:        baseURL + "/",
		LastMod:    time.Now().Format("2006-01-02"),
		ChangeFreq: "daily",
		Priority:   1.0,
	})

	// 2. Published Posts
	var posts []models.Post
	database.DB.Where("is_draft = ? AND type = ?", false, "post").Select("slug, updated_at").Find(&posts)
	for _, post := range posts {
		urls = append(urls, SitemapURL{
			Loc:        fmt.Sprintf("%s/post/%s", baseURL, post.Slug),
			LastMod:    post.UpdatedAt.Format("2006-01-02"),
			ChangeFreq: "weekly",
			Priority:   0.8,
		})
	}

	// 3. Published Pages
	var pages []models.Post
	database.DB.Where("is_draft = ? AND type = ?", false, "page").Select("slug, updated_at").Find(&pages)
	for _, page := range pages {
		urls = append(urls, SitemapURL{
			Loc:        fmt.Sprintf("%s/page/%s", baseURL, page.Slug),
			LastMod:    page.UpdatedAt.Format("2006-01-02"),
			ChangeFreq: "monthly",
			Priority:   0.6,
		})
	}

	// 4. Categories
	var categories []models.Category
	database.DB.Find(&categories)
	for _, cat := range categories {
		var lastModPost time.Time
		err := database.DB.Model(&models.Post{}).
			Where("category_id = ? AND is_draft = ? AND type = ?", cat.ID, false, "post").
			Select("MAX(updated_at)").
			Scan(&lastModPost).Error

		lastModStr := time.Now().Format("2006-01-02")
		if err == nil && !lastModPost.IsZero() {
			lastModStr = lastModPost.Format("2006-01-02")
		}

		urls = append(urls, SitemapURL{
			Loc:        fmt.Sprintf("%s/category/%s", baseURL, cat.Slug),
			LastMod:    lastModStr,
			ChangeFreq: "weekly",
			Priority:   0.5,
		})
	}

	// 5. Tags
	var tags []models.Tag
	database.DB.Find(&tags)
	for _, tag := range tags {
		var lastModPost time.Time
		err := database.DB.Model(&models.Post{}).
			Joins("JOIN post_tags ON post_tags.post_id = posts.id").
			Where("post_tags.tag_id = ? AND posts.is_draft = ? AND posts.type = ?", tag.ID, false, "post").
			Select("MAX(posts.updated_at)").
			Scan(&lastModPost).Error

		lastModStr := time.Now().Format("2006-01-02")
		if err == nil && !lastModPost.IsZero() {
			lastModStr = lastModPost.Format("2006-01-02")
		}

		urls = append(urls, SitemapURL{
			Loc:        fmt.Sprintf("%s/tag/%s", baseURL, tag.Slug),
			LastMod:    lastModStr,
			ChangeFreq: "weekly",
			Priority:   0.4,
		})
	}

	urlset := SitemapURLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	output, err := xml.MarshalIndent(urlset, "", "  ")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error generating sitemap")
	}

	c.Set("Content-Type", "application/xml")
	return c.SendString(xml.Header + string(output))
}
