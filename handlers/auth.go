// handlers/auth.go
package handlers

import (
	"time"

	"go-lightblog/config"
	"go-lightblog/database"
	"go-lightblog/models"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func ProcessLogin(c *fiber.Ctx) error {
	urlToken := c.Params("token")
	validToken := models.GetSetting(database.DB, "login_token", "admin")

	// Cegah akses form POST dari sumber eksternal tanpa token
	if urlToken != validToken {
		return c.Status(fiber.StatusNotFound).SendString("404 Not Found")
	}

	username := c.FormValue("username")
	password := c.FormValue("password")
	remember := c.FormValue("remember") // Akan berisi "true" jika dicentang

	// Helper function untuk merender ulang halaman login dengan error
	renderError := func(errMsg string) error {
		return c.Status(fiber.StatusUnauthorized).Render("login", fiber.Map{
			"Error":    errMsg,
			"Username": username, // Kembalikan username agar tidak perlu mengetik ulang
		})
	}

	// Cari user di database
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		// Idealnya, kembalikan flash message error ke halaman login
		return renderError("Username atau Password salah.")
	}

	// Bandingkan Password bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return renderError("Username atau Password salah.")
	}

	// Ambil Session
	sess, err := config.SessStore.Get(c)
	if err != nil {
		return renderError("Terjadi kesalahan pada sistem sesi. Silakan coba lagi.")
	}

	// Simpan data login di dalam sesi
	sess.Set("is_logged_in", true)
	sess.Set("user_id", user.ID)

	// === TAMBAHAN BARU UNTUK RBAC ===
	sess.Set("name", user.Name) // Simpan nama tampilan ke sesi
	sess.Set("role", user.Role) // Simpan role (admin/editor) ke sesi
	// ================================

	// Logika Remember Me
	if remember == "true" {
		// Set sesi bertahan lebih lama (misal 7 hari)
		sess.SetExpiry(7 * 24 * time.Hour)
	} else {
        // Biarkan mengikuti Expiration default dari config (2 Jam)
        // Kita bisa reset expiry jika ingin memastikan
        sess.SetExpiry(2 * time.Hour)
    }

	// Simpan perubahan sesi
	if err := sess.Save(); err != nil {
		return renderError("Gagal menyimpan sesi login.")
	}

	// Redirect ke Dashboard (rutenya akan kita buat nanti)
	return c.Redirect("/dashboard")
}

func ProcessLogout(c *fiber.Ctx) error {
	sess, err := config.SessStore.Get(c)
	if err != nil {
		return c.Redirect("/")
	}
	validToken := models.GetSetting(database.DB, "login_token", "admin")

	// Hancurkan sesi
	sess.Destroy()

	return c.Redirect("/login-" + validToken)
}