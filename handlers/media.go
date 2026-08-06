package handlers

import (
	"go-lightblog/utils"

	"github.com/gofiber/fiber/v2"
)

func ImageThumbProxy(c *fiber.Ctx) error {
	src := c.Query("src")
	w := c.QueryInt("w", 600)

	thumbPath, err := utils.ResizeAndCacheThumbnail(src, w)
	if err != nil {
		switch err.Error() {
		case "access denied":
			return c.Status(fiber.StatusForbidden).SendString("Access denied")
		case "original image not found":
			return c.Status(fiber.StatusNotFound).SendString("Original image not found")
		case "image too large", "unsupported image format":
			return c.Status(fiber.StatusBadRequest).SendString("Invalid image parameters")
		default:
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to process thumbnail")
		}
	}

	c.Set("Cache-Control", "public, max-age=31536000")
	return c.SendFile(thumbPath)
}