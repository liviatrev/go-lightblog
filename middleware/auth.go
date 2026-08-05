package middleware

import (
	"strings"

	"go-lightblog/config"
	"go-lightblog/database"
	"go-lightblog/models"

	"github.com/gofiber/fiber/v2"
)

// RequireLogin adalah middleware untuk memproteksi rute yang butuh autentikasi
func RequireLogin(c *fiber.Ctx) error {
	sess, err := config.SessStore.Get(c)
	loginToken := models.GetSetting(database.DB, "login_token", "admin")
	if err != nil {
		// Jika terjadi error baca sesi, anggap tidak terautentikasi
		return c.Redirect("/login-" + loginToken)
	}

	// Mengecek apakah key "is_logged_in" ada di dalam sesi
	if sess.Get("is_logged_in") == nil {
		// Sesi tidak ditemukan atau kedaluwarsa
		return c.Redirect("/login-" + loginToken)
	}

	// Sesi valid, lanjutkan ke handler tujuan
	return c.Next()
}

func RequireAdmin(c *fiber.Ctx) error {
	// Panggil config.SessStore sesuai strukturmu
	sess, err := config.SessStore.Get(c)
	loginToken := models.GetSetting(database.DB, "login_token", "admin")
	if err != nil {
		return c.Redirect("/login-" + loginToken)
	}

	// Pastikan role-nya adalah admin
	role, ok := sess.Get("role").(string)
	if !ok || role != "admin" {
		// Jika editor, lemparkan kembali ke dashboard
		return c.Redirect("/dashboard")
	}

	return c.Next()
}

// RequireRole membatasi akses berdasarkan Role pengguna
func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// PERBAIKAN: Gunakan konversi tipe aman untuk mencegah server Panic
		roleVal, ok := c.Locals("user_role").(string) 
		if !ok || roleVal == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Akses ditolak. Sesi atau peran tidak terdeteksi.",
			})
		}

		// Periksa apakah role user ada di daftar role yang diizinkan
		isAllowed := false
		for _, role := range allowedRoles {
			if roleVal == role {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Akses dilarang. Anda tidak memiliki izin untuk tindakan ini.",
			})
		}

		return c.Next()
	}
}

// RequireAuth memastikan klien memiliki token API yang valid
func RequireAuth(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: Token tidak ditemukan atau format salah",
		})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	var user models.User
	
	// PERBAIKAN: Menggunakan First() agar GORM melempar error jika token tidak ada
	if err := database.DB.Where("api_key = ?", tokenString).
		Select("role", "id").First(&user).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Token API tidak valid atau tidak terdaftar",
			})
		}

	// Simpan data role untuk dibaca oleh RequireRole
	c.Locals("user_role", user.Role)
	c.Locals("user_id", user.ID)
	return c.Next()
}