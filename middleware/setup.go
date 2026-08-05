package middleware

import (
	"go-lightblog/config"

	"github.com/gofiber/fiber/v2"
)

func CheckSetup(c *fiber.Ctx) error {
	// Pengecualian: Izinkan akses ke file statis (CSS/JS/Img)
	// Kita tidak ingin halaman setup tampil tanpa styling CSS
	if len(c.Path()) >= 7 && c.Path()[:7] == "/public" {
		return c.Next()
	}

	// Jika CMS sudah di-setup, ATAU user memang sedang berada di jalur /setup
	// Biarkan lewat.
	if config.AppSetupCompleted || c.Path() == "/setup" || c.Path() == "/setup/process" {
		// Proteksi ekstra: Jika sudah setup tapi memaksa akses /setup, lempar ke depan
		if config.AppSetupCompleted && (c.Path() == "/setup" || c.Path() == "/setup/process") {
			return c.Redirect("/")
		}
		
		return c.Next()
	}

	// Jika belum setup dan mengakses rute selain /setup, lempar secara paksa
	return c.Redirect("/setup")
}