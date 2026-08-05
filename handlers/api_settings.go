package handlers

import (
	"go-lightblog/database"
	"go-lightblog/models"

	"github.com/gofiber/fiber/v2"
)

// ApiGetSettings mengambil data pengaturan global CMS (Hanya untuk Admin)
func ApiGetSettings(c *fiber.Ctx) error {
	var settings []models.Setting

	// Ambil seluruh baris pengaturan dari tabel
	if err := database.DB.Find(&settings).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data pengaturan.",
		})
	}

	// Rakit slice of struct menjadi map tunggal agar format JSON-nya rapi untuk frontend
	// Hasilnya nanti: {"site_title": "Blog Livia", "ai_prompt": "..."}
	settingsMap := make(map[string]string)
	for _, s := range settings {
		settingsMap[s.Key] = s.Value
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    settingsMap,
	})
}

// ApiUpdateSettings memperbarui data pengaturan secara parsial (mendukung partial update)
func ApiUpdateSettings(c *fiber.Ctx) error {
	// 1. Tangkap payload dinamis menjadi Map
	var payload map[string]string

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format JSON tidak valid.",
			"error":   err.Error(),
		})
	}

	// 2. Daftar putih (Allowlist) key yang diizinkan untuk diubah via API.
	// Ini adalah benteng keamanan agar user tidak mengirim key sampah (misal: "hacked_status": "yes")
	allowedKeys := map[string]bool{
		"site_title":           true,
		"site_description":     true,
		"site_keywords":        true,
		"upload_mode":          true,
		"imagekit_private_key": true,
		"imagekit_folder":      true,
		"remark42_url":         true,
		"remark42_site_id":     true,
		"enable_gemini":        true,
		"gemini_api_key":       true,
		"gemini_model":         true,
		"login_token":          true,
	}

	var updatedKeys []string

	// 3. Looping hanya pada data yang dikirim oleh klien
	for key, value := range payload {
		// Cek apakah key tersebut valid dan ada di daftar putih
		if allowedKeys[key] {
			// Lakukan Upsert (Update jika ada, Insert jika belum ada)
			setting := models.Setting{Key: key, Value: value}
			if err := database.DB.Save(&setting).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Gagal menyimpan pengaturan dengan key: " + key,
					"error":   err.Error(),
				})
			}
			updatedKeys = append(updatedKeys, key)
		}
	}

	// 4. Jika tidak ada key valid yang diperbarui
	if len(updatedKeys) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Tidak ada data pengaturan valid yang diperbarui. Pastikan key JSON sesuai.",
		})
	}

	// 5. Kembalikan respons sukses beserta daftar key yang berhasil disentuh
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":      true,
		"message":      "Pengaturan berhasil diperbarui secara parsial.",
		"updated_keys": updatedKeys,
	})
}