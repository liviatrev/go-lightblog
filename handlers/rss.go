// handlers/rss.go
package handlers

import (
	"encoding/xml"
	"fmt"
	"time"
	"strings"

	"go-lightblog/database"
	"go-lightblog/models"
	"go-lightblog/utils"

	"github.com/gofiber/fiber/v2"
)

// HandleRSS renders a dynamic RSS 2.0 feed of the 20 latest published posts.
// It is served at both /feed.xml and /rss.xml.
func HandleRSS(c *fiber.Ctx) error {
	baseURL := utils.GetSiteURL()
	if baseURL == "" {
		baseURL = c.BaseURL()
	}

	// Normalize base URL (strip trailing slash)
	for len(baseURL) > 1 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	// Fetch the 20 latest published posts with their relationships
	var posts []models.Post
	database.DB.Where("is_draft = ? AND type = ?", false, "post").
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		Order("created_at desc").
		Limit(20).
		Find(&posts)

	// Build channel metadata from settings
	siteTitle := utils.GetSiteTitle()
	siteDesc := models.GetSetting(database.DB, "site_description", "A minimal and fast blog powered by Go Fiber.")
	language := models.GetSetting(database.DB, "site_language", "en")

	// lastBuildDate: use the newest post's CreatedAt, or now if no posts
	lastBuild := time.Now()
	if len(posts) > 0 {
		lastBuild = posts[0].CreatedAt
	}

	selfURL := fmt.Sprintf("%s%s", baseURL, c.Path())

	channel := models.RSSChannel{
		Title:         siteTitle,
		Link:          baseURL + "/",
		Description:   siteDesc,
		Language:      language,
		LastBuildDate: lastBuild.Format(time.RFC1123Z),
		AtomLink:      models.AtomLink{
			Href: selfURL,
			Rel:  "self",
			Type: "application/rss+xml",
		},
		Items:         make([]models.RSSItem, 0, len(posts)),
	}

	// Build items for each post
	for _, post := range posts {
		postLink := fmt.Sprintf("%s/post/%s", baseURL, post.Slug)

		item := models.RSSItem{
			Title:       post.Title,
			Link:        postLink,
			Description: post.MetaDescription,
			PubDate:     post.CreatedAt.Format(time.RFC1123Z),
			GUID:        postLink,
			Author:      post.Author.Name,
			Categories:  make([]string, 0),
		}

		seenCategories := make(map[string]bool)
		// Add main category as a <category> element
		if post.Category.Name != "" {
			lowerName := strings.ToLower(post.Category.Name)
			seenCategories[lowerName] = true
			item.Categories = append(item.Categories, post.Category.Name)
		}

		// Add all tags as <category> elements
		for _, tag := range post.Tags {
			if tag.Name == "" {
				continue
			}
			lowerName := strings.ToLower(tag.Name)

			if !seenCategories[lowerName] {
				seenCategories[lowerName] = true
				item.Categories = append(item.Categories, tag.Name)
			}
		}

		channel.Items = append(channel.Items, item)
	}

	feed := models.RSS{
		Version: "2.0",
		Xmlns:   "http://www.w3.org/2005/Atom",
		Channel: channel,
	}

	output, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error generating RSS feed")
	}

	c.Set("Content-Type", "application/rss+xml; charset=utf-8")
	return c.SendString(xml.Header + string(output))
}