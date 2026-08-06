package handlers

import (
	"go-lightblog/utils"

	"github.com/gofiber/fiber/v2"
)

func GenerateSEO(c *fiber.Ctx) error {
	var req utils.SEORequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request format"})
	}

	seoData, err := utils.ProcessSEOInternal(req.Content)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(seoData)
}