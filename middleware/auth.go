package middleware

import (
	"strings"

	"go-lightblog/config"
	"go-lightblog/database"
	"go-lightblog/models"
	"go-lightblog/utils"

	"github.com/gofiber/fiber/v2"
)

// RequireLogin is middleware to protect routes that need authentication.
// Unauthenticated users are served a 404 page instead of being redirected
// to the hidden login URL, so the login token is never leaked.
func RequireLogin(c *fiber.Ctx) error {
	sess, err := config.SessStore.Get(c)
	if err != nil {
		// If error reading session, consider unauthenticated
		return utils.RenderNotFound(c)
	}

	// Check if "is_logged_in" key exists in session
	if sess.Get("is_logged_in") == nil {
		// Session not found or expired
		return utils.RenderNotFound(c)
	}

	// Session valid, continue to destination handler
	return c.Next()
}

func RequireAdmin(c *fiber.Ctx) error {
	// Call config.SessStore according to your structure
	sess, err := config.SessStore.Get(c)
	if err != nil {
		return utils.RenderNotFound(c)
	}

	// Make sure role is admin
	role, ok := sess.Get("role").(string)
	if !ok || role != "admin" {
		// If editor, redirect back to dashboard
		return c.Redirect("/dashboard")
	}

	return c.Next()
}

// RequireRole restricts access based on user Role
func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// FIX: Use safe type conversion to prevent server Panic
		roleVal, ok := c.Locals("user_role").(string) 
		if !ok || roleVal == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Access denied. Session or role not detected.",
			})
		}

		// Check if user role is in the allowed roles list
		isAllowed := false
		for _, role := range allowedRoles {
			if roleVal == role {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access forbidden. You do not have permission for this action.",
			})
		}

		return c.Next()
	}
}

// RequireAuth ensures client has a valid API token
func RequireAuth(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: Token not found or invalid format",
		})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	var user models.User
	
	// FIX: Using First() so GORM throws error if token doesn't exist
	if err := database.DB.Where("api_key = ?", tokenString).
		Select("role", "id").First(&user).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: API Token invalid or not registered",
			})
		}

	// Store role data for RequireRole to read
	c.Locals("user_role", user.Role)
	c.Locals("user_id", user.ID)
	return c.Next()
}
