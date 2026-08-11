package handlers

import (
	"strings"

	"go-lightblog/database"
	"go-lightblog/models"
	"go-lightblog/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func ApiGetPosts(c *fiber.Ctx) error {
	// Pagination support to prevent loading the entire table at once
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var total int64
	database.DB.Model(&models.Post{}).Count(&total)

	var posts []models.Post

	if err := database.DB.Preload("Category").Preload("Tags").Preload("Author").
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch article data",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    posts,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

func ApiCreatePost(c *fiber.Ctx) error {
	title := c.FormValue("title")
	content := c.FormValue("content")
	categoryID := c.FormValue("category_id")
	tagsInput := c.FormValue("tags")
	isDraft := c.FormValue("isDraft") == "true"

	if strings.TrimSpace(title) == "" || strings.TrimSpace(content) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Title and content cannot be empty",
		})
	}

	post := models.Post{
		Title:      title,
		Slug:       utils.GenerateSlug(title),
		Content:    content,
		IsDraft:    isDraft,
		CategoryID: utils.ParseUint(categoryID),
	}

	post.AuthorID = c.Locals("user_id").(uint)

	coverURL, err := utils.ProcessUpload(c, "cover")
	if err == nil && coverURL != "" {
		post.CoverImage = coverURL
	}

	enableGemini := models.GetSetting(database.DB, "enable_gemini", "no")
	if enableGemini == "yes" {
		geminiKey := models.GetSetting(database.DB, "gemini_api_key", "")
		if geminiKey != "" {
			if post.MetaTitle == "" || post.MetaDescription == "" || post.TargetKeyword == "" {
				seoResult, err := utils.ProcessSEOInternal(post.Content)
				if err == nil && seoResult != nil {
					if post.MetaTitle == "" {
						post.MetaTitle = seoResult.MetaTitle
					}
					if post.MetaDescription == "" {
						post.MetaDescription = seoResult.MetaDescription
					}
					if post.TargetKeyword == "" {
						post.TargetKeyword = seoResult.TargetKeyword
					}
				}
			}
		}
	} else {
		post.MetaTitle = c.FormValue("seo_title")
		post.MetaDescription = c.FormValue("seo_description")
		post.TargetKeyword = c.FormValue("seo_keywords")
	}

	if err := database.DB.Create(&post).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to save article",
			"error":   err.Error(),
		})
	}

	if tagsInput != "" {
		tagIDs := strings.Split(tagsInput, ",")
		var tags []models.Tag
		database.DB.Where("id IN ?", tagIDs).Find(&tags)
		database.DB.Model(&post).Association("Tags").Append(tags)
	}

	database.DB.Preload("Category").
		Preload("Tags").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, name, role")
		}).
		First(&post, post.ID)

	// Purge Cloudflare cache for this post, its category, tags, and homepage.
	go func() {
		if err := utils.PurgePostCache(post.ID); err != nil {
			utils.LogCloudflareError("API create post", err)
		}
		utils.PurgeSitemapCache()
	}()

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Article published successfully!",
		"data":    post,
	})
}

func ApiUpdatePost(c *fiber.Ctx) error {
	id := c.Params("id")
	var post models.Post

	if err := database.DB.First(&post, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Article not found",
		})
	}

	oldSlug := post.Slug

	// 1. Prepare container for partial data
	// Columns adjusted to GORM database column names (snake_case)
	updateData := make(map[string]interface{})

	// 2. Read multipart form to detect which keys were sent
	form, err := c.MultipartForm()
	if err == nil && form != nil {
		if values, exists := form.Value["title"]; exists && len(values) > 0 {
			newTitle := values[0]
			if post.Title != newTitle {
				updateData["title"] = newTitle
				updateData["slug"] = utils.GenerateUniqueSlug(newTitle, post.ID)
			}
		}
		if values, exists := form.Value["content"]; exists && len(values) > 0 {
			updateData["content"] = values[0]
		}
		if values, exists := form.Value["isDraft"]; exists && len(values) > 0 {
			updateData["is_draft"] = values[0] == "true"
		}
		if values, exists := form.Value["category_id"]; exists && len(values) > 0 {
			updateData["category_id"] = values[0]
		}
		if values, exists := form.Value["seo_title"]; exists && len(values) > 0 {
			updateData["meta_title"] = values[0]
		}
		if values, exists := form.Value["seo_description"]; exists && len(values) > 0 {
			updateData["meta_description"] = values[0]
		}
		if values, exists := form.Value["seo_keywords"]; exists && len(values) > 0 {
			updateData["target_keyword"] = values[0]
		}
	}

	// 3. Partial Cover Image Handling
	// ProcessUpload will fail (err != nil) if user doesn't include "cover" file
	coverURL, errUpload := utils.ProcessUpload(c, "cover")
	if errUpload == nil && coverURL != "" {
		updateData["cover_image"] = coverURL
	}

	// 4. Execute Partial Update ONLY if there is data sent
	if len(updateData) > 0 {
		if err := database.DB.Model(&post).Updates(updateData).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to apply partial update",
			})
		}
	}

	// 5. Tags Handling (Many-to-Many Pivot)
	// Since this is a separate table relationship, must be handled outside Update map
	if form != nil {
		if values, exists := form.Value["tags"]; exists && len(values) > 0 {
			tagsInput := values[0]
			var tags []models.Tag

			// If user sends empty tags (e.g. wants to remove all tags)
			if tagsInput == "" {
				database.DB.Model(&post).Association("Tags").Clear()
			} else {
				tagIDs := strings.Split(tagsInput, ",")
				database.DB.Where("id IN ?", tagIDs).Find(&tags)
				database.DB.Model(&post).Association("Tags").Replace(tags)
			}
		}
	}

	// 6. Reload Data for JSON response
	database.DB.Preload("Category").
		Preload("Tags").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, name, role")
		}).
		First(&post, post.ID)

	if oldSlug != post.Slug {
		database.DB.Create(&models.SlugRedirect{
			OldSlug: oldSlug,
			NewSlug: post.Slug,
		})
	}

	// Purge Cloudflare cache for this post, its category, tags, and homepage.
	go func() {
		if err := utils.PurgePostCache(post.ID); err != nil {
			utils.LogCloudflareError("API edit post", err)
		}
		utils.PurgeSitemapCache()
	}()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Article updated partially successfully!",
		"data":    post,
	})
}

func ApiDeletePost(c *fiber.Ctx) error {
	id := c.Params("id")
	var post models.Post

	// Check if article with that ID exists (preload Category and Tags for cache purging)
	if err := database.DB.Preload("Category").Preload("Tags").First(&post, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Article not found",
		})
	}

	// Execute Soft Delete
	// GORM will automatically fill deleted_at column and NOT remove the row from SQLite
	if err := database.DB.Delete(&post).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete article",
		})
	}

	// Purge Cloudflare cache for this post, its category, tags, and homepage.
	go func() {
		if err := utils.PurgePostCacheByPost(post); err != nil {
			utils.LogCloudflareError("API delete post", err)
		}
		utils.PurgeSitemapCache()
	}()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Article moved to trash successfully",
	})
}
