package handlers

import (
	"go-lightblog/database"
	"go-lightblog/models"
	"go-lightblog/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type TagInput struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func ApiGetTags(c *fiber.Ctx) error {
	var tags []models.Tag

	if err := database.DB.Find(&tags).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data tag",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    tags,
	})
}

func ApiCreateTag(c *fiber.Ctx) error {
	var input TagInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format JSON tidak valid",
		})
	}

	if strings.TrimSpace(input.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Nama tag tidak boleh kosong",
		})
	}

	if strings.TrimSpace(input.Slug) == "" {
		input.Slug = utils.GenerateSlug(input.Name)
	}

	tag := models.Tag{
		Name: input.Name,
		Slug: input.Slug,
	}

	if err := database.DB.Create(&tag).Error; err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyimpan tag. Pastikan Slug unik.",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Tag berhasil dibuat",
		"data":    tag,
	})
}

// ApiUpdateTag - Memperbarui tag yang ada
func ApiUpdateTag(c *fiber.Ctx) error {
	id := c.Params("id")
	var input TagInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format JSON tidak valid",
		})
	}

	var tag models.Tag
	if err := database.DB.First(&tag, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Tag tidak ditemukan",
		})
	}

	if input.Name != "" {
		tag.Name = input.Name
	}
	if input.Slug != "" {
		tag.Slug = input.Slug
	}

	if err := database.DB.Save(&tag).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memperbarui tag",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Tag berhasil diperbarui",
		"data":    tag,
	})
}

// ApiDeleteTag - Menghapus tag
func ApiDeleteTag(c *fiber.Ctx) error {
	id := c.Params("id")

	var tag models.Tag
	if err := database.DB.First(&tag, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Tag tidak ditemukan",
		})
	}

	if err := database.DB.Delete(&tag).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus tag",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Tag berhasil dihapus",
	})
}