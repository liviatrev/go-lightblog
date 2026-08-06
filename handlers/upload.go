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