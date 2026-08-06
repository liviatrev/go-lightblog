package middleware

import (
	"go-lightblog/config"

	"github.com/gofiber/fiber/v2"
)

func CheckSetup(c *fiber.Ctx) error {
	// Exception: Allow access to static files (CSS/JS/Img)
	// We don't want the setup page to appear without CSS styling
	if len(c.Path()) >= 7 && c.Path()[:7] == "/public" {
		return c.Next()
	}

	// If CMS is already set up, OR user is currently on /setup path
	// Let them through.
	if config.AppSetupCompleted || c.Path() == "/setup" || c.Path() == "/setup/process" {
		// Extra protection: If already set up but forcing /setup access, redirect forward
		if config.AppSetupCompleted && (c.Path() == "/setup" || c.Path() == "/setup/process") {
			return c.Redirect("/")
		}
		
		return c.Next()
	}

	// If not set up and accessing route other than /setup, force redirect
	return c.Redirect("/setup")
}
