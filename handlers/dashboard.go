// handlers/dashboard.go
package handlers

import (
	"go-lightblog/database"
	"go-lightblog/models"
	"go-lightblog/utils"

	"github.com/gofiber/fiber/v2"
)

func DashboardView(c *fiber.Ctx) error {
	var totalPosts int64
	var recentPosts []models.Post

	su := utils.GetSessionUser(c)

	database.DB.Model(&models.Post{}).Count(&totalPosts)
	database.DB.Order("updated_at desc").Limit(5).Find(&recentPosts)

	return c.Render("dashboard/index", fiber.Map{
		"Title":       "Dashboard",
		"HeaderTitle": "Dashboard Overview",
		"ActiveMenu":  "dashboard",
		"TotalPosts":  totalPosts,
		"RecentPosts": recentPosts,
		"UserName":    su.UserName,
		"UserRole":    su.UserRole,
	}, "layouts/main")
}