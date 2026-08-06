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
			"message": "Failed to fetch tag data",
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
			"message": "Invalid JSON format",
		})
	}

	if strings.TrimSpace(input.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Tag name cannot be empty",
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
			"message": "Failed to save tag. Make sure Slug is unique.",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Tag created successfully",
		"data":    tag,
	})
}

// ApiUpdateTag - Updates an existing tag
func ApiUpdateTag(c *fiber.Ctx) error {
	id := c.Params("id")
	var input TagInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid JSON format",
		})
	}

	var tag models.Tag
	if err := database.DB.First(&tag, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Tag not found",
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
			"message": "Failed to update tag",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Tag updated successfully",
		"data":    tag,
	})
}

// ApiDeleteTag - Deletes a tag
func ApiDeleteTag(c *fiber.Ctx) error {
	id := c.Params("id")

	var tag models.Tag
	if err := database.DB.First(&tag, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Tag not found",
		})
	}

	if err := database.DB.Delete(&tag).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete tag",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Tag deleted successfully",
	})
}