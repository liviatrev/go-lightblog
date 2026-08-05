package handlers

import (
	"strings"

	"go-lightblog/database"
	"go-lightblog/models"
	"go-lightblog/utils"

	"github.com/gofiber/fiber/v2"
)

type CategoryInput struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func ApiGetCategories(c *fiber.Ctx) error {
	var categories []models.Category

	if err := database.DB.Find(&categories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data kategori",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    categories,
	})
}

func ApiCreateCategory(c *fiber.Ctx) error {
	var input CategoryInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format JSON tidak valid",
		})
	}

	if strings.TrimSpace(input.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Nama kategori tidak boleh kosong",
		})
	}

	if strings.TrimSpace(input.Slug) == "" {
		input.Slug = utils.GenerateSlug(input.Name)
	}

	category := models.Category{
		Name: input.Name,
		Slug: input.Slug,
	}

	if err := database.DB.Create(&category).Error; err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyimpan kategori. Pastikan Slug unik.",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Kategori berhasil dibuat",
		"data":    category,
	})
}

// ApiUpdateCategory - Memperbarui kategori yang ada
func ApiUpdateCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	var input CategoryInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format JSON tidak valid",
		})
	}

	var category models.Category
	// Gunakan First() agar melempar error jika ID tidak ditemukan
	if err := database.DB.First(&category, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Kategori tidak ditemukan",
		})
	}

	// Update data (hanya jika ada nilainya)
	if input.Name != "" {
		category.Name = input.Name
	}
	if input.Slug != "" {
		category.Slug = input.Slug
	}

	if err := database.DB.Save(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memperbarui kategori",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Kategori berhasil diperbarui",
		"data":    category,
	})
}

// ApiDeleteCategory - Menghapus kategori
func ApiDeleteCategory(c *fiber.Ctx) error {
	id := c.Params("id")

	var category models.Category
	if err := database.DB.First(&category, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Kategori tidak ditemukan",
		})
	}

	if err := database.DB.Delete(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus kategori",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Kategori berhasil dihapus",
	})
}