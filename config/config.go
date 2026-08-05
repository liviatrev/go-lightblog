// config/config.go
package config

import (
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
)

// AppSetupCompleted menyimpan status apakah CMS sudah di-setup
var AppSetupCompleted bool = false

// SessStore adalah variabel global untuk manajemen session
var SessStore *session.Store

func InitSession() {
	SessStore = session.New(session.Config{
		// Kedaluwarsa default jika tidak mencentang "Remember Me" (misal: 2 jam)
		Expiration: 2 * time.Hour, 
		// Cookie HTTP Only dan SameSite untuk keamanan XSS & CSRF dasar
		CookieHTTPOnly: true,
		CookieSameSite: "Strict",
	})
}