package handlers

import (
	"math"
	"strings"

	"go-lightblog/database"
	"go-lightblog/models"
	"go-lightblog/utils"

	"github.com/gofiber/fiber/v2"
)

func Home(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := 6
	offset := (page - 1) * limit

	var totalPosts int64
	database.DB.Model(&models.Post{}).Where("is_draft = ? AND type = ?", false, "post").Count(&totalPosts)

	totalPages := int(math.Ceil(float64(totalPosts) / float64(limit)))

	var posts []models.Post

	database.DB.Where("is_draft = ? AND type = ?", false, "post").
		Preload("Category").
		Preload("Tags").
		Preload("Author").
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&posts)

	siteTitle := utils.GetSiteTitle()

	siteDesc := models.GetSetting(database.DB, "site_description", "A minimal and fast blog powered by Go Fiber.")
	siteKeywords := models.GetSetting(database.DB, "site_keywords", "blog, go, fiber, lightblog")
	siteHeadline := models.GetSetting(database.DB, "site_headline", "Explore Articles")
	siteTagline := models.GetSetting(database.DB, "site_tagline", "A collection of the latest writings, notes, and insights.")

	data := utils.GetNavbarData()

	data["SiteTitle"] = siteTitle
	data["SiteDescription"] = siteDesc
	data["SiteKeywords"] = siteKeywords
	data["SiteHeadline"] = siteHeadline
	data["SiteTagline"] = siteTagline
	data["Posts"] = posts
	data["CurrentPage"] = page
	data["TotalPages"] = totalPages
	data["HasPrev"] = page > 1
	data["HasNext"] = page < totalPages
	data["PrevPage"] = page - 1
	data["NextPage"] = page + 1

	return c.Render("home", data, "layouts/public")
}

// ReadPost displays the full content of an article by slug
func ReadPost(c *fiber.Ctx) error {
	slug := c.Params("slug")
	remark42URL := models.GetSetting(database.DB, "remark42_url", "http://127.0.0.1:8080")
	remark42SiteID := models.GetSetting(database.DB, "remark42_site_id", "lightblog-utama")

	var post models.Post

	// Find article by slug and ensure it's not a draft
	if err := database.DB.Where("slug = ? AND is_draft = ?", slug, false).
		Preload("Category").
		Preload("Tags").
		Preload("Author").
		First(&post).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Article not found or not published.")
	}

	baseURL := c.BaseURL()
	metaImageURL := post.CoverImage
	if metaImageURL != "" && !strings.HasPrefix(metaImageURL, "http") {
		metaImageURL = baseURL + metaImageURL
	}

	data := utils.GetNavbarData()

	data["SiteTitle"] = utils.GetSiteTitle()
	data["SiteDescription"] = post.MetaDescription
	data["SiteKeywords"] = post.TargetKeyword
	data["Post"] = post
	data["MetaImageURL"] = metaImageURL
	data["Remark42URL"] = remark42URL
	data["Remark42SiteID"] = remark42SiteID

	return c.Render("post", data, "layouts/public")
}

// SearchPosts handles article search by keyword
func SearchPosts(c *fiber.Ctx) error {
	query := strings.TrimSpace(c.Query("q"))
	page := c.QueryInt("page", 1)
	limit := 6
	offset := (page - 1) * limit

	var posts []models.Post
	var totalPosts int64

	if query != "" {
		// Search by Title, Meta Description, or Content
		// Only search type 'post' and not draft
		searchTerm := "%" + query + "%"
		dbQuery := database.DB.Model(&models.Post{}).
			Where("is_draft = ? AND type = ?", false, "post").
			Where("title LIKE ? OR meta_description LIKE ? OR content LIKE ?", searchTerm, searchTerm, searchTerm)

		// Count total search results
		dbQuery.Count(&totalPosts)

		// Get search result data
		dbQuery.Preload("Category").
			Preload("Author").
			Order("created_at desc").
			Limit(limit).
			Offset(offset).
			Find(&posts)
	}

	totalPages := int(math.Ceil(float64(totalPosts) / float64(limit)))

	data := utils.GetNavbarData()
	data["SiteTitle"] = utils.GetSiteTitle()
	data["Query"] = query
	data["Posts"] = posts
	data["TotalPosts"] = totalPosts
	data["CurrentPage"] = page
	data["TotalPages"] = totalPages
	data["HasPrev"] = page > 1
	data["HasNext"] = page < totalPages
	data["PrevPage"] = page - 1
	data["NextPage"] = page + 1

	return c.Render("search", data, "layouts/public")
}

// CategoryPosts displays a list of articles by Category
func CategoryPosts(c *fiber.Ctx) error {
	slug := c.Params("slug")
	page := c.QueryInt("page", 1)
	limit := 6
	offset := (page - 1) * limit

	// Find category by slug
	var category models.Category
	if err := database.DB.Where("slug = ?", slug).First(&category).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Category not found")
	}

	var posts []models.Post
	var total int64

	dbQuery := database.DB.Model(&models.Post{}).Where("category_id = ? AND is_draft = ? AND type = ?", category.ID, false, "post")
	dbQuery.Count(&total)

	dbQuery.Preload("Category").Preload("Author").Preload("Tags").
		Order("created_at desc").Limit(limit).Offset(offset).Find(&posts)

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	data := utils.GetNavbarData()
	data["SiteTitle"] = utils.GetSiteTitle()
	data["ArchiveTitle"] = "Category: " + category.Name
	data["Posts"] = posts
	data["CurrentPage"] = page
	data["TotalPages"] = totalPages
	data["HasPrev"] = page > 1
	data["HasNext"] = page < totalPages
	data["PrevPage"] = page - 1
	data["NextPage"] = page + 1
	// We use an extra parameter for pagination link to match current URL
	data["PaginationURL"] = "/category/" + slug

	return c.Render("archive", data, "layouts/public")
}

// TagPosts displays a list of articles by Tag
func TagPosts(c *fiber.Ctx) error {
	slug := c.Params("slug")
	page := c.QueryInt("page", 1)
	limit := 6
	offset := (page - 1) * limit

	var tag models.Tag
	if err := database.DB.Where("slug = ?", slug).First(&tag).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Tag not found")
	}

	var posts []models.Post
	var total int64

	// Join many-to-many relationship table (assume default GORM table name is post_tags)
	dbQuery := database.DB.Model(&models.Post{}).
		Joins("JOIN post_tags ON post_tags.post_id = posts.id").
		Where("post_tags.tag_id = ? AND posts.is_draft = ? AND posts.type = ?", tag.ID, false, "post")

	dbQuery.Count(&total)
	dbQuery.Preload("Category").Preload("Author").Preload("Tags").
		Order("posts.created_at desc").Limit(limit).Offset(offset).Find(&posts)

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	data := utils.GetNavbarData()
	data["SiteTitle"] = utils.GetSiteTitle()
	data["ArchiveTitle"] = "Tag: #" + tag.Name
	data["Posts"] = posts
	data["CurrentPage"] = page
	data["TotalPages"] = totalPages
	data["HasPrev"] = page > 1
	data["HasNext"] = page < totalPages
	data["PrevPage"] = page - 1
	data["NextPage"] = page + 1
	data["PaginationURL"] = "/tag/" + slug

	return c.Render("archive", data, "layouts/public")
}

// AuthorPosts displays a list of articles by Author
func AuthorPosts(c *fiber.Ctx) error {
	id := c.Params("id") // Using Author ID as URL parameter
	page := c.QueryInt("page", 1)
	limit := 6
	offset := (page - 1) * limit

	var author models.User
	if err := database.DB.Where("id = ?", id).First(&author).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Author not found")
	}

	var posts []models.Post
	var total int64

	dbQuery := database.DB.Model(&models.Post{}).Where("author_id = ? AND is_draft = ? AND type = ?", author.ID, false, "post")
	dbQuery.Count(&total)

	dbQuery.Preload("Category").Preload("Author").Preload("Tags").
		Order("created_at desc").Limit(limit).Offset(offset).Find(&posts)

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	data := utils.GetNavbarData()
	data["SiteTitle"] = utils.GetSiteTitle()
	data["ArchiveTitle"] = "Author: " + author.Name
	data["Posts"] = posts
	data["CurrentPage"] = page
	data["TotalPages"] = totalPages
	data["HasPrev"] = page > 1
	data["HasNext"] = page < totalPages
	data["PrevPage"] = page - 1
	data["NextPage"] = page + 1
	data["PaginationURL"] = "/author/" + id

	return c.Render("archive", data, "layouts/public")
}
