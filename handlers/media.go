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
		case "akses ditolak":
			return c.Status(fiber.StatusForbidden).SendString("Akses ditolak")
		case "gambar asli tidak ditemukan":
			return c.Status(fiber.StatusNotFound).SendString("Gambar asli tidak ditemukan")
		case "gambar terlalu besar", "format gambar tidak didukung":
			return c.Status(fiber.StatusBadRequest).SendString("Parameter gambar tidak valid")
		default:
			return c.Status(fiber.StatusInternalServerError).SendString("Gagal memproses thumbnail")
		}
	}

	c.Set("Cache-Control", "public, max-age=31536000")
	return c.SendFile(thumbPath)
}