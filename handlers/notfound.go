// handlers/notfound.go
package handlers

import (
	"go-lightblog/utils"

	"github.com/gofiber/fiber/v2"
)

// NotFound renders a custom 404 page using the public layout.
// It is used both for unmatched routes and for protected admin
// routes accessed without authentication (to avoid leaking the
// hidden login URL).
func NotFound(c *fiber.Ctx) error {
	return utils.RenderNotFound(c)
}