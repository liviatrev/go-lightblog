// handlers/settings.go
package handlers

import (
	"go-lightblog/database"
	"go-lightblog/models"
	"go-lightblog/utils"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func SettingsView(c *fiber.Ctx) error {
	siteTitle := utils.GetSiteTitle()

	su := utils.GetSessionUser(c)
	msg := c.Query("msg")

	uploadMode := models.GetSetting(database.DB, "upload_mode", "local")
	imagekitKey := models.GetSetting(database.DB, "imagekit_private_key", "")
	imagekitFolder := models.GetSetting(database.DB, "imagekit_folder", "/lightblog")
	remark42URL := models.GetSetting(database.DB, "remark42_url", "http://127.0.0.1:8080")
	remark42SiteID := models.GetSetting(database.DB, "remark42_site_id", "lightblog-utama")
	enableGemini := models.GetSetting(database.DB, "enable_gemini", "no")
	geminiKey := models.GetSetting(database.DB, "gemini_api_key", "")
	geminiModel := models.GetSetting(database.DB, "gemini_model", "gemini-flash-latest")
	siteDesc := models.GetSetting(database.DB, "site_description", "A minimal and fast blog powered by Go Fiber.")
	siteKeywords := models.GetSetting(database.DB, "site_keywords", "blog, go, fiber, lightblog")
	siteHeadline := models.GetSetting(database.DB, "site_headline", "Explore Articles")
	siteTagline := models.GetSetting(database.DB, "site_tagline", "A collection of the latest writings, notes, and insights.")
	loginToken := models.GetSetting(database.DB, "login_token", "admin")

	return c.Render("dashboard/settings", fiber.Map{
		"Title":           "Settings",
		"HeaderTitle":     "System Settings",
		"SiteTitle":       siteTitle,
		"ActiveMenu":      "settings",
		"Message":         msg,
		"UserName":        su.UserName,
		"UserRole":        su.UserRole,
		"UploadMode":      uploadMode,
		"ImagekitKey":     imagekitKey,
		"ImagekitFolder":  imagekitFolder,
		"Remark42URL":     remark42URL,
		"Remark42SiteID":  remark42SiteID,
		"GeminiKey":       geminiKey,
		"GeminiModel":     geminiModel,
		"SiteDesc":        siteDesc,
		"SiteKeywords":    siteKeywords,
		"SiteHeadline":	   siteHeadline,
		"SiteTagline":	   siteTagline,
		"EnableGemini":    enableGemini,
		"LoginToken":      loginToken,
	}, "layouts/main")
}

func ProcessUpdateSettings(c *fiber.Ctx) error {
	newTitle := c.FormValue("site_title")

	if newTitle == "" {
		return c.Redirect("/settings?msg=error_empty")
	}

	keys := []string{"site_title", "site_description", "site_keywords", "site_headline", "site_tagline"}
	for _, key := range keys {
		database.DB.Save(&models.Setting{
			Key:   key,
			Value: c.FormValue(key),
		})
	}

	return c.Redirect("/settings?msg=success")
}

func ProcessUpdatePassword(c *fiber.Ctx) error {
	newPassword := c.FormValue("newPassword")
	confirmPassword := c.FormValue("confirmPassword")

	if len(newPassword) < 8 || newPassword != confirmPassword {
		return c.Redirect("/settings?msg=error_password")
	}

	su := utils.GetSessionUser(c)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Redirect("/settings?msg=error_hash")
	}

	if err := database.DB.Model(&models.User{}).Where("id = ?", su.UserID).Update("password", string(hashedPassword)).Error; err != nil {
		return c.Redirect("/settings?msg=error_db")
	}

	currentToken := models.GetSetting(database.DB, "login_token", "admin")
	newToken := c.FormValue("loginToken")
	if newToken != "" {
		database.DB.Save(&models.Setting{
			Key:   "login_token",
			Value: newToken,
		})
		currentToken = newToken
	}

	su.Session.Destroy()
	return c.Redirect("/login-" + currentToken)
}

func ProcessUpdateIntegrations(c *fiber.Ctx) error {
	keys := []string{"upload_mode", "imagekit_private_key", "imagekit_folder", "remark42_url", "remark42_site_id", "enable_gemini", "gemini_api_key", "gemini_model"}

	for _, key := range keys {
		database.DB.Save(&models.Setting{
			Key:   key,
			Value: c.FormValue(key),
		})
	}

	return c.Redirect("/settings?msg=success")
}