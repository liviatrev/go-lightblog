// handlers/auth.go
package handlers

import (
	"time"

	"go-lightblog/config"
	"go-lightblog/database"
	"go-lightblog/models"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func ProcessLogin(c *fiber.Ctx) error {
	urlToken := c.Params("token")
	validToken := models.GetSetting(database.DB, "login_token", "admin")

	// Prevent POST access from external sources without token
	if urlToken != validToken {
		return c.Status(fiber.StatusNotFound).SendString("404 Not Found")
	}

	username := c.FormValue("username")
	password := c.FormValue("password")
	remember := c.FormValue("remember") // Will contain "true" if checked

	// Helper function to re-render the login page with error
	renderError := func(errMsg string) error {
		return c.Status(fiber.StatusUnauthorized).Render("login", fiber.Map{
			"Error":    errMsg,
			"Username": username, // Return username so user doesn't have to retype
		})
	}

	// Find user in database
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		// Ideally, return a flash message error to the login page
		return renderError("Username or Password is incorrect.")
	}

	// Compare bcrypt Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return renderError("Username or Password is incorrect.")
	}

	// Get Session
	sess, err := config.SessStore.Get(c)
	if err != nil {
		return renderError("An error occurred with the session system. Please try again.")
	}

	// Store login data in session
	sess.Set("is_logged_in", true)
	sess.Set("user_id", user.ID)

	// === NEW FOR RBAC ===
	sess.Set("name", user.Name) // Store display name in session
	sess.Set("role", user.Role) // Store role (admin/editor) in session
	// ================================

	// Remember Me Logic
	if remember == "true" {
		// Set session to last longer (e.g. 7 days)
		sess.SetExpiry(7 * 24 * time.Hour)
	} else {
        // Follow the default Expiration from config (2 hours)
        // We can reset expiry if we want to ensure
        sess.SetExpiry(2 * time.Hour)
    }

	// Save session changes
	if err := sess.Save(); err != nil {
		return renderError("Failed to save login session.")
	}

	// Redirect to Dashboard (route will be created later)
	return c.Redirect("/dashboard")
}

func ProcessLogout(c *fiber.Ctx) error {
	sess, err := config.SessStore.Get(c)
	if err != nil {
		return c.Redirect("/")
	}
	validToken := models.GetSetting(database.DB, "login_token", "admin")

	// Destroy session
	sess.Destroy()

	return c.Redirect("/login-" + validToken)
}