// handlers/upload.go
package handlers

import (
	"go-lightblog/utils"

	"github.com/gofiber/fiber/v2"
)

func UploadImage(c *fiber.Ctx) error {
	url, err := utils.ProcessUpload(c, "image")

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to upload to CDN: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"url":     url,
	})
}

// ApiUploadImage handles image uploads (cover or in-content) via the
// admin REST API using Bearer token authentication. It accepts the file
// from either the "image" field (in-content images from the editor) or
// the "cover" field (post cover images), so a single endpoint can serve
// both use cases.
func ApiUploadImage(c *fiber.Ctx) error {
	// Try the "image" field first (in-content image), then fall back to
	// the "cover" field (cover image).
	url, err := utils.ProcessUpload(c, "image")
	if err != nil {
		url, err = utils.ProcessUpload(c, "cover")
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to upload to CDN: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"url":     url,
	})
}