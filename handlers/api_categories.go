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
			"message": "Failed to fetch category data",
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
			"message": "Invalid JSON format",
		})
	}

	if strings.TrimSpace(input.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Category name cannot be empty",
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
			"message": "Failed to save category. Make sure Slug is unique.",
			"error":   err.Error(),
		})
	}

	// Purge Cloudflare cache for this category and homepage.
	go func() {
		if err := utils.PurgeTaxonomyCache(category.Slug, "category"); err != nil {
			utils.LogCloudflareError("API create category", err)
		}
		utils.PurgeSitemapCache()
	}()

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Category created successfully",
		"data":    category,
	})
}

// ApiUpdateCategory - Updates an existing category
func ApiUpdateCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	var input CategoryInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid JSON format",
		})
	}

	var category models.Category
	// Use First() to throw error if ID not found
	if err := database.DB.First(&category, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Category not found",
		})
	}

	// Update data (only if there is a value)
	if input.Name != "" {
		category.Name = input.Name
	}
	if input.Slug != "" {
		category.Slug = input.Slug
	}

	if err := database.DB.Save(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update category",
		})
	}

	// Purge Cloudflare cache for this category and homepage.
	go func() {
		if err := utils.PurgeTaxonomyCache(category.Slug, "category"); err != nil {
			utils.LogCloudflareError("API edit category", err)
		}
		utils.PurgeSitemapCache()
	}()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Category updated successfully",
		"data":    category,
	})
}

// ApiDeleteCategory - Deletes a category
func ApiDeleteCategory(c *fiber.Ctx) error {
	id := c.Params("id")

	var category models.Category
	if err := database.DB.First(&category, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Category not found",
		})
	}

	if err := database.DB.Delete(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete category",
		})
	}

	go utils.PurgeSitemapCache()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Category deleted successfully",
	})
}
