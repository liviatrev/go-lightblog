// handlers/indexnow.go
package handlers

import (
	"go-lightblog/database"
	"go-lightblog/models"

	"github.com/gofiber/fiber/v2"
)

// IndexNowKeyFile serves the IndexNow key verification file.
// The route is /{indexnow_key}.txt and its body is the key value.
// If the key in the URL doesn't match the configured key, return 404.
func IndexNowKeyFile(c *fiber.Ctx) error {
	key := c.Params("key")
	configuredKey := models.GetSetting(database.DB, "indexnow_key", "")

	if key == "" || configuredKey == "" || key != configuredKey {
		return c.Status(fiber.StatusNotFound).SendString("404 Not Found")
	}

	c.Set("Content-Type", "text/plain; charset=utf-8")
	return c.SendString(configuredKey)
}