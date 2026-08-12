// handlers/posts.go
package handlers

import (
	"math"
	"strconv"
	"strings"

	"go-lightblog/database"
	"go-lightblog/models"
	"go-lightblog/utils"

	"github.com/gofiber/fiber/v2"
)

func ListPosts(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	var totalPosts int64
	database.DB.Model(&models.Post{}).Count(&totalPosts)

	var posts []models.Post
	database.DB.Order("created_at desc").Limit(limit).Offset(offset).Find(&posts)

	totalPages := int(math.Ceil(float64(totalPosts) / float64(limit)))

	su := utils.GetSessionUser(c)
	return c.Render("dashboard/posts_list", fiber.Map{
		"Title":       "Post List",
		"HeaderTitle": "Manage Posts",
		"ActiveMenu":  "posts",
		"Posts":       posts,
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"HasPrev":     page > 1,
		"HasNext":     page < totalPages,
		"PrevPage":    page - 1,
		"NextPage":    page + 1,
		"UserName":    su.UserName,
		"UserRole":    su.UserRole,
	}, "layouts/main")
}

func CreatePostView(c *fiber.Ctx) error {
	var categories []models.Category
	var tags []models.Tag
	database.DB.Find(&categories)
	database.DB.Find(&tags)

	su := utils.GetSessionUser(c)
	enableGemini := models.GetSetting(database.DB, "enable_gemini", "no")

	return c.Render("dashboard/posts_create", fiber.Map{
		"Title":        "Write New Post",
		"HeaderTitle":  "Write New Article",
		"HeaderPostForm": "Write Article",
		"ActiveMenu":   "posts",
		"Categories":   categories,
		"Tags":         tags,
		"EnableGemini": enableGemini,
		"UserName":     su.UserName,
		"UserRole":     su.UserRole,
	}, "layouts/main")
}

func ProcessCreatePost(c *fiber.Ctx) error {
	su := utils.GetSessionUser(c)

	title := c.FormValue("title")
	content := c.FormValue("content")
	postType := c.FormValue("type")
	if postType == "" {
		postType = "post"
	}
	isDraft := c.FormValue("isDraft") == "true"

	if title == "" || content == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Title and Content cannot be empty")
	}

	categoryID := c.FormValue("category_id")
	coverImage := c.FormValue("cover_image")
	metaTitle := c.FormValue("meta_title")
	metaDesc := c.FormValue("meta_description")
	targetKeyword := c.FormValue("target_keyword")

	slug := utils.GenerateUniqueSlug(title)

	// Auto-generate a cover image when the user did not upload one.
	if strings.TrimSpace(coverImage) == "" {
		if coverURL, genErr := utils.GenerateCoverImage(slug, title); genErr == nil && coverURL != "" {
			coverImage = coverURL
		} else if genErr != nil {
			utils.LogCloudflareError("auto cover create", genErr)
		}
	}

	tagsArgs := c.Request().PostArgs().PeekMulti("tags[]")
	var selectedTags []models.Tag
	for _, tagIDBytes := range tagsArgs {
		id, _ := strconv.Atoi(string(tagIDBytes))
		selectedTags = append(selectedTags, models.Tag{ID: uint(id)})
	}

	post := models.Post{
		Title:           title,
		Slug:            slug,
		Content:         content,
		IsDraft:         isDraft,
		CoverImage:      coverImage,
		MetaTitle:       metaTitle,
		MetaDescription: metaDesc,
		TargetKeyword:   targetKeyword,
		AuthorID:        su.UserID,
		CategoryID:      utils.ParseUint(categoryID),
		Tags:            selectedTags,
		Type:            postType,
	}

	if err := database.DB.Create(&post).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to save post")
	}

	// Purge Cloudflare cache for this post, its category, tags, and homepage.
	// Non-blocking: wraps in goroutine so response time isn't affected by the purge.
	go func() {
		if err := utils.PurgePostCache(post.ID); err != nil {
			utils.LogCloudflareError("create post", err)
		}
		utils.PurgeSitemapCache()
	}()

	// Submit the post URL to IndexNow when it is published (not a draft).
	if !post.IsDraft {
		go utils.SubmitIndexNowURL(post)
	}

	return c.Redirect("/posts")
}

func EditPostView(c *fiber.Ctx) error {
	id := c.Params("id")
	var post models.Post

	su := utils.GetSessionUser(c)

	if err := database.DB.First(&post, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Post not found")
	}

	if err := database.DB.Preload("Tags").First(&post, id).Error; err != nil {
		return c.Redirect("/posts")
	}

	var categories []models.Category
	var tags []models.Tag
	database.DB.Find(&categories)
	database.DB.Find(&tags)

	selectedTagsMap := make(map[uint]bool)
	for _, t := range post.Tags {
		selectedTagsMap[t.ID] = true
	}
	enableGemini := models.GetSetting(database.DB, "enable_gemini", "no")

	return c.Render("dashboard/posts_edit", fiber.Map{
		"Title":           "Edit Post",
		"HeaderTitle":	   "Edit Article",
		"HeaderPostForm":  "Edit: " + post.Title,
		"ActiveMenu":      "posts",
		"Post":            post,
		"Categories":      categories,
		"Tags":            tags,
		"SelectedTagsMap": selectedTagsMap,
		"UserName":        su.UserName,
		"UserRole":        su.UserRole,
		"EnableGemini":    enableGemini,
		"Type":            post.Type,
	}, "layouts/main")
}

func ProcessEditPost(c *fiber.Ctx) error {
	id := c.Params("id")
	var post models.Post

	if err := database.DB.First(&post, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Post not found")
	}

	oldSlug := post.Slug

	title := c.FormValue("title")
	content := c.FormValue("content")

	if title == "" || content == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Title and Content cannot be empty")
	}

	if post.Title != title {
		post.Slug = utils.GenerateUniqueSlug(title, post.ID)
	}

	post.Title = title
	post.Content = content
	post.IsDraft = c.FormValue("isDraft") == "true"

	// Keep existing cover unless the user provided a new one or the post
	// has never had a cover (in which case auto-generate from the slug).
	incomingCover := c.FormValue("cover_image")
	if strings.TrimSpace(incomingCover) != "" {
		post.CoverImage = incomingCover
	} else if strings.TrimSpace(post.CoverImage) == "" {
		if coverURL, genErr := utils.GenerateCoverImage(post.Slug, title); genErr == nil && coverURL != "" {
			post.CoverImage = coverURL
		} else if genErr != nil {
			utils.LogCloudflareError("auto cover edit", genErr)
		}
	}

	post.MetaTitle = c.FormValue("meta_title")
	post.MetaDescription = c.FormValue("meta_description")
	post.TargetKeyword = c.FormValue("target_keyword")
	postType := c.FormValue("type")
	if postType != "" {
		post.Type = postType
	}

	post.CategoryID = utils.ParseUint(c.FormValue("category_id"))

	if err := database.DB.Save(&post).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to update post")
	}

	if oldSlug != post.Slug {
		database.DB.Create(&models.SlugRedirect{
			OldSlug: oldSlug,
			NewSlug: post.Slug,
		})
	}

	tagsArgs := c.Request().PostArgs().PeekMulti("tags[]")
	var selectedTags []models.Tag
	for _, tagIDBytes := range tagsArgs {
		tagID, _ := strconv.Atoi(string(tagIDBytes))
		selectedTags = append(selectedTags, models.Tag{ID: uint(tagID)})
	}

	database.DB.Model(&post).Association("Tags").Replace(selectedTags)

	// Purge Cloudflare cache for this post, its category, tags, and homepage.
	go func() {
		if err := utils.PurgePostCache(post.ID); err != nil {
			utils.LogCloudflareError("edit post", err)
		}
		utils.PurgeSitemapCache()
	}()

	// Submit the post URL to IndexNow when it is published (not a draft).
	if !post.IsDraft {
		go utils.SubmitIndexNowURL(post)
	}

	return c.Redirect("/posts")
}

// DeletePost deletes an article from the database (soft delete)
func DeletePost(c *fiber.Ctx) error {
	id := c.Params("id")

	// Get post with category and tags for cache purging
	var post models.Post
	if err := database.DB.Preload("Category").Preload("Tags").First(&post, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Post not found")
	}

	// Soft delete (sets DeletedAt, keeps record in database)
	if err := database.DB.Delete(&post).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to delete post")
	}

	// Purge Cloudflare cache using already-loaded post data
	go func() {
		if err := utils.PurgePostCacheByPost(post); err != nil {
			utils.LogCloudflareError("delete post", err)
		}
		utils.PurgeSitemapCache()
	}()

	// Submit the deleted post URL to IndexNow so search engines remove it.
	go utils.SubmitIndexNowURL(post)

	return c.Redirect("/posts")
}
