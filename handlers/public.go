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

	siteDesc := models.GetSetting(database.DB, "site_description", "Blog minimalis dan cepat bertenaga Go Fiber.")
	siteKeywords := models.GetSetting(database.DB, "site_keywords", "blog, go, fiber, lightblog")

	data := utils.GetNavbarData()

	data["SiteTitle"] = siteTitle
	data["SiteDescription"] = siteDesc
	data["SiteKeywords"] = siteKeywords
	data["Posts"] = posts
	data["CurrentPage"] = page
	data["TotalPages"] = totalPages
	data["HasPrev"] = page > 1
	data["HasNext"] = page < totalPages
	data["PrevPage"] = page - 1
	data["NextPage"] = page + 1

	return c.Render("home", data, "layouts/public")
}

// ReadPost menampilkan isi penuh satu artikel berdasarkan slug
func ReadPost(c *fiber.Ctx) error {
	slug := c.Params("slug")
	remark42URL := models.GetSetting(database.DB, "remark42_url", "http://127.0.0.1:8080")
	remark42SiteID := models.GetSetting(database.DB, "remark42_site_id", "lightblog-utama")

	var post models.Post

	// Cari artikel berdasarkan slug dan pastikan bukan draft
	if err := database.DB.Where("slug = ? AND is_draft = ?", slug, false).
		Preload("Category").
		Preload("Tags").
		Preload("Author").
		First(&post).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Artikel tidak ditemukan atau belum dipublikasikan.")
	}

	baseURL := c.BaseURL()
	metaImageURL := post.CoverImage
	if metaImageURL != "" && !strings.HasPrefix(metaImageURL, "http") {
		metaImageURL = baseURL + metaImageURL // Gabungkan: https://domainmu.com + /public/uploads/gambar.jpg
	}

	// return c.Render("post", fiber.Map{
	// 	"SiteTitle": siteTitle,
	// 	"SiteDescription": siteDesc,
	// 	"Post":      post,
	// 	"MetaImageURL":    metaImageURL,
	// 	"Remark42URL":    remark42URL,
	// 	"Remark42SiteID": remark42SiteID,
	// }, "layouts/public")

	data := utils.GetNavbarData()

	data["SiteTitle"] = post.Title + " - " + utils.GetSiteTitle()
	data["SiteDescription"] = post.MetaDescription
	data["SiteKeywords"] = post.TargetKeyword
	data["Post"] = post
	data["MetaImageURL"] = metaImageURL
	data["Remark42URL"] = remark42URL
	data["Remark42SiteID"] = remark42SiteID

	return c.Render("post", data, "layouts/public")
}

// SearchPosts menangani pencarian artikel berdasarkan kata kunci
func SearchPosts(c *fiber.Ctx) error {
	query := strings.TrimSpace(c.Query("q"))
	page := c.QueryInt("page", 1)
	limit := 6
	offset := (page - 1) * limit

	var posts []models.Post
	var totalPosts int64

	if query != "" {
		// Cari berdasarkan Judul, Meta Description, atau Konten
		// Hanya cari tipe 'post' dan bukan draft
		searchTerm := "%" + query + "%"
		dbQuery := database.DB.Model(&models.Post{}).
			Where("is_draft = ? AND type = ?", false, "post").
			Where("title LIKE ? OR meta_description LIKE ? OR content LIKE ?", searchTerm, searchTerm, searchTerm)

		// Hitung total hasil pencarian
		dbQuery.Count(&totalPosts)

		// Ambil data hasil pencarian
		dbQuery.Preload("Category").
			Preload("Author").
			Order("created_at desc").
			Limit(limit).
			Offset(offset).
			Find(&posts)
	}

	totalPages := int(math.Ceil(float64(totalPosts) / float64(limit)))

	data := utils.GetNavbarData()
	data["SiteTitle"] = "Pencarian: " + query + " - " + utils.GetSiteTitle()
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

// CategoryPosts menampilkan daftar artikel berdasarkan Kategori
func CategoryPosts(c *fiber.Ctx) error {
	slug := c.Params("slug")
	page := c.QueryInt("page", 1)
	limit := 6
	offset := (page - 1) * limit

	// Cari kategori berdasarkan slug
	var category models.Category
	if err := database.DB.Where("slug = ?", slug).First(&category).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Kategori tidak ditemukan")
	}

	var posts []models.Post
	var total int64

	dbQuery := database.DB.Model(&models.Post{}).Where("category_id = ? AND is_draft = ? AND type = ?", category.ID, false, "post")
	dbQuery.Count(&total)

	dbQuery.Preload("Category").Preload("Author").Preload("Tags").
		Order("created_at desc").Limit(limit).Offset(offset).Find(&posts)

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	data := utils.GetNavbarData()
	data["SiteTitle"] = "Kategori: " + category.Name + " - " + utils.GetSiteTitle()
	data["ArchiveTitle"] = "Kategori: " + category.Name
	data["Posts"] = posts
	data["CurrentPage"] = page
	data["TotalPages"] = totalPages
	data["HasPrev"] = page > 1
	data["HasNext"] = page < totalPages
	data["PrevPage"] = page - 1
	data["NextPage"] = page + 1
	// Kita gunakan parameter ekstra untuk link pagination agar sesuai dengan URL saat ini
	data["PaginationURL"] = "/category/" + slug

	return c.Render("archive", data, "layouts/public")
}

// TagPosts menampilkan daftar artikel berdasarkan Tag
func TagPosts(c *fiber.Ctx) error {
	slug := c.Params("slug")
	page := c.QueryInt("page", 1)
	limit := 6
	offset := (page - 1) * limit

	var tag models.Tag
	if err := database.DB.Where("slug = ?", slug).First(&tag).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Tag tidak ditemukan")
	}

	var posts []models.Post
	var total int64

	// Join tabel relasi many-to-many (asumsi nama tabel default GORM adalah post_tags)
	dbQuery := database.DB.Model(&models.Post{}).
		Joins("JOIN post_tags ON post_tags.post_id = posts.id").
		Where("post_tags.tag_id = ? AND posts.is_draft = ? AND posts.type = ?", tag.ID, false, "post")

	dbQuery.Count(&total)
	dbQuery.Preload("Category").Preload("Author").Preload("Tags").
		Order("posts.created_at desc").Limit(limit).Offset(offset).Find(&posts)

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	data := utils.GetNavbarData()
	data["SiteTitle"] = "Tag: " + tag.Name + " - " + utils.GetSiteTitle()
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

// AuthorPosts menampilkan daftar artikel berdasarkan Penulis
func AuthorPosts(c *fiber.Ctx) error {
	id := c.Params("id") // Kita gunakan ID Penulis sebagai parameter URL
	page := c.QueryInt("page", 1)
	limit := 6
	offset := (page - 1) * limit

	var author models.User
	if err := database.DB.Where("id = ?", id).First(&author).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Penulis tidak ditemukan")
	}

	var posts []models.Post
	var total int64

	dbQuery := database.DB.Model(&models.Post{}).Where("author_id = ? AND is_draft = ? AND type = ?", author.ID, false, "post")
	dbQuery.Count(&total)

	dbQuery.Preload("Category").Preload("Author").Preload("Tags").
		Order("created_at desc").Limit(limit).Offset(offset).Find(&posts)

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	data := utils.GetNavbarData()
	data["SiteTitle"] = "Penulis: " + author.Name + " - " + utils.GetSiteTitle()
	data["ArchiveTitle"] = "Penulis: " + author.Name
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
