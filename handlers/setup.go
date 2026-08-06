package handlers

import (
	"go-lightblog/config"
	"go-lightblog/database"
	"go-lightblog/models"
	"go-lightblog/utils"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func SetupView(c *fiber.Ctx) error {
	var count int64
	database.DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		loginToken := models.GetSetting(database.DB, "login_token", "admin")
		return c.Redirect("/login-" + loginToken)
	}

	return c.Render("setup", fiber.Map{
		"Title": "LightBlog Installation",
	}, "layouts/main")
}

func SetupProcess(c *fiber.Ctx) error {
	var count int64
	database.DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		return c.Redirect("/login")
	}

	username := c.FormValue("username")
	password := c.FormValue("password")
	name := c.FormValue("name")
	siteTitle := c.FormValue("site_title")

	if username == "" || password == "" || siteTitle == "" || name == "" {
		return c.Render("setup", fiber.Map{
			"Title": "LightBlog Installation",
			"Error": "All fields (Username, Password, Name, Site Title) are required",
		}, "layouts/main")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).SendString("Internal Server Error (Bcrypt)")
	}

	user := models.User{
		Username: username,
		Password: string(hashedPassword),
		Name:     name,
		Role:     "admin",
		APIKey:   utils.GenerateAPIKey(),
	}
	database.DB.Create(&user)

	setting := models.Setting{
		Key:   "site_title",
		Value: siteTitle,
	}
	database.DB.Create(&setting)

	loginToken := utils.GenerateRandomHex(5)

	database.DB.Save(&models.Setting{
		Key:   "login_token",
		Value: loginToken,
	})

	config.AppSetupCompleted = true

	return c.Redirect("/login-" + loginToken)
}