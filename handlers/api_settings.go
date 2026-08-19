package handlers

import (
	"go-lightblog/database"
	"go-lightblog/models"

	"github.com/gofiber/fiber/v2"
)

// ApiGetSettings gets global CMS settings data (Admin only)
func ApiGetSettings(c *fiber.Ctx) error {
	var settings []models.Setting

	// Get all setting rows from the table
	if err := database.DB.Find(&settings).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch settings data.",
		})
	}

	// Assemble slice of struct into a single map so JSON format is clean for frontend
	// Result will be: {"site_title": "Blog Livia", "ai_prompt": "..."}
	settingsMap := make(map[string]string)
	for _, s := range settings {
		settingsMap[s.Key] = s.Value
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    settingsMap,
	})
}

// ApiUpdateSettings updates settings data partially (supports partial update)
func ApiUpdateSettings(c *fiber.Ctx) error {
	// 1. Capture dynamic payload into Map
	var payload map[string]string

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid JSON format.",
			"error":   err.Error(),
		})
	}

	// 2. Allowlist of keys allowed to be changed via API.
	// This is a security fortress so users can't send junk keys (e.g. "hacked_status": "yes")
	allowedKeys := map[string]bool{
		"site_title":           true,
		"site_description":     true,
		"site_keywords":        true,
		"site_headline":		 true,
		"site_tagline":			 true,
		"upload_mode":          true,
		"imagekit_private_key": true,
		"imagekit_folder":      true,
		"remark42_url":         true,
		"remark42_site_id":     true,
		"enable_gemini":        true,
		"gemini_api_key":       true,
		"gemini_model":         true,
		"login_token":          true,
		"enable_cloudflare":    true,
		"cloudflare_api_key":   true,
		"cloudflare_zone_id":   true,
		"site_url":             true,
		"public_theme":         true,
		"indexnow":             true,
		"indexnow_key":         true,
		"header_script":        true,
		"footer_script":        true,
	}

	var updatedKeys []string

	// 3. Loop only on data sent by client
	for key, value := range payload {
		// Check if the key is valid and in the allowlist
		if allowedKeys[key] {
			// Perform Upsert (Update if exists, Insert if not)
			setting := models.Setting{Key: key, Value: value}
			if err := database.DB.Save(&setting).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Failed to save setting with key: " + key,
					"error":   err.Error(),
				})
			}
			updatedKeys = append(updatedKeys, key)
		}
	}

	// 4. If no valid key was updated
	if len(updatedKeys) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "No valid settings data was updated. Make sure JSON keys match.",
		})
	}

	// 5. Return success response with list of keys that were touched
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":      true,
		"message":      "Settings updated partially successfully.",
		"updated_keys": updatedKeys,
	})
}