package handlers

import (
	"go-lightblog/database"
	"go-lightblog/models"
	"go-lightblog/utils"

	"github.com/gofiber/fiber/v2"
)

// ==========================================
// KATEGORI
// ==========================================

func CategoryList(c *fiber.Ctx) error {
	var categories []models.Category
	database.DB.Find(&categories)

	su := utils.GetSessionUser(c)
	return c.Render("dashboard/categories", fiber.Map{
		"Title":       "Kategori",
		"HeaderTitle": "Manajemen Kategori",
		"ActiveMenu":  "categories",
		"Categories":  categories,
		"UserName":    su.UserName,
		"UserRole":    su.UserRole,
	}, "layouts/main")
}

func CategoryCreate(c *fiber.Ctx) error {
	name := c.FormValue("name")
	if name != "" {
		database.DB.Create(&models.Category{
			Name: name,
			Slug: utils.GenerateSlug(name),
		})
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
		"Title":       "Tag",
		"HeaderTitle": "Manajemen Tag",
		"ActiveMenu":  "tags",
		"Tags":        tags,
		"UserName":    su.UserName,
		"UserRole":    su.UserRole,
	}, "layouts/main")
}

func TagCreate(c *fiber.Ctx) error {
	name := c.FormValue("name")
	if name != "" {
		database.DB.Create(&models.Tag{
			Name: name,
			Slug: utils.GenerateSlug(name),
		})
	}
	return c.Redirect("/tags")
}

func TagDelete(c *fiber.Ctx) error {
	database.DB.Delete(&models.Tag{}, c.Params("id"))
	return c.Redirect("/tags")
}