package handlers

import (
	"go-lightblog/database"
	"go-lightblog/models"
	"go-lightblog/utils"

	"github.com/gofiber/fiber/v2"
)

// ==========================================
// CATEGORY
// ==========================================

func CategoryList(c *fiber.Ctx) error {
	var categories []models.Category
	database.DB.Find(&categories)

	su := utils.GetSessionUser(c)
	return c.Render("dashboard/categories", fiber.Map{
		"Title":       "Categories",
		"HeaderTitle": "Manage Categories",
		"ActiveMenu":  "categories",
		"Categories":  categories,
		"UserName":    su.UserName,
		"UserRole":    su.UserRole,
	}, "layouts/main")
}

func CategoryCreate(c *fiber.Ctx) error {
	name := c.FormValue("name")
	if name != "" {
		category := models.Category{
			Name: name,
			Slug: utils.GenerateSlug(name),
		}
		database.DB.Create(&category)

		// Purge Cloudflare cache for this category and homepage.
		go func() {
			if err := utils.PurgeTaxonomyCache(category.Slug, "category"); err != nil {
				utils.LogCloudflareError("create category", err)
			}
		}()
	}
	return c.Redirect("/categories")
}

func CategoryDelete(c *fiber.Ctx) error {
	database.DB.Delete(&models.Category{}, c.Params("id"))
	return c.Redirect("/categories")
}

// ==========================================
// TAG
// ==========================================

func TagList(c *fiber.Ctx) error {
	var tags []models.Tag
	database.DB.Find(&tags)

	su := utils.GetSessionUser(c)
	return c.Render("dashboard/tags", fiber.Map{
		"Title":       "Tags",
		"HeaderTitle": "Manage Tags",
		"ActiveMenu":  "tags",
		"Tags":        tags,
		"UserName":    su.UserName,
		"UserRole":    su.UserRole,
	}, "layouts/main")
}

func TagCreate(c *fiber.Ctx) error {
	name := c.FormValue("name")
	if name != "" {
		tag := models.Tag{
			Name: name,
			Slug: utils.GenerateSlug(name),
		}
		database.DB.Create(&tag)

		// Purge Cloudflare cache for this tag and homepage.
		go func() {
			if err := utils.PurgeTaxonomyCache(tag.Slug, "tag"); err != nil {
				utils.LogCloudflareError("create tag", err)
			}
		}()
	}
	return c.Redirect("/tags")
}

func TagDelete(c *fiber.Ctx) error {
	database.DB.Delete(&models.Tag{}, c.Params("id"))
	return c.Redirect("/tags")
}