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
	var posts []models.Post

	if err := database.DB.Preload("Category").Preload("Tags").Preload("Author").Find(&posts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data artikel",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    posts,
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
			"message": "Judul dan konten tidak boleh kosong",
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
			"message": "Gagal menyimpan artikel",
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

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Artikel berhasil diterbitkan!",
		"data":    post,
	})
}

func ApiUpdatePost(c *fiber.Ctx) error {
	id := c.Params("id")
	var post models.Post

	if err := database.DB.First(&post, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Artikel tidak ditemukan",
		})
	}

	// 1. Siapkan keranjang penampung data parsial
	// Kolom disesuaikan dengan nama kolom GORM di database (snake_case)
	updateData := make(map[string]interface{})

	// 2. Baca form multipart untuk mendeteksi kunci apa saja yang dikirim
	form, err := c.MultipartForm()
	if err == nil && form != nil {
		if values, exists := form.Value["title"]; exists && len(values) > 0 {
			updateData["title"] = values[0]
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

	// 3. Penanganan Gambar Kover Parsial
	// ProcessUpload akan gagal (err != nil) jika user tidak menyertakan file "cover"
	coverURL, errUpload := utils.ProcessUpload(c, "cover")
	if errUpload == nil && coverURL != "" {
		updateData["cover_image"] = coverURL
	}

	// 4. Eksekusi Partial Update HANYA jika ada data yang dikirim
	if len(updateData) > 0 {
		if err := database.DB.Model(&post).Updates(updateData).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal menerapkan pembaruan parsial",
			})
		}
	}

	// 5. Penanganan Tags (Many-to-Many Pivot)
	// Karena ini relasi tabel terpisah, harus ditangani di luar Update map
	if form != nil {
		if values, exists := form.Value["tags"]; exists && len(values) > 0 {
			tagsInput := values[0]
			var tags []models.Tag

			// Jika user mengirim tags kosong (misal ingin menghapus semua tag)
			if tagsInput == "" {
				database.DB.Model(&post).Association("Tags").Clear()
			} else {
				tagIDs := strings.Split(tagsInput, ",")
				database.DB.Where("id IN ?", tagIDs).Find(&tags)
				database.DB.Model(&post).Association("Tags").Replace(tags)
			}
		}
	}

	// 6. Reload Data untuk kembalian JSON
	database.DB.Preload("Category").
		Preload("Tags").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, username, name, role")
		}).
		First(&post, post.ID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Artikel berhasil diperbarui secara parsial!",
		"data":    post,
	})
}

func ApiDeletePost(c *fiber.Ctx) error {
	id := c.Params("id")
	var post models.Post

	// Cek apakah artikel dengan ID tersebut ada
	if err := database.DB.First(&post, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Artikel tidak ditemukan",
		})
	}

	// Eksekusi Soft Delete
	// GORM otomatis akan mengisi kolom deleted_at dan TIDAK menghapus baris dari SQLite
	if err := database.DB.Delete(&post).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus artikel",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Artikel berhasil dipindahkan ke tong sampah",
	})
}
