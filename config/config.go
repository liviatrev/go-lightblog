// config/config.go
package config

import (
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
)

// AppSetupCompleted stores whether the CMS has been set up
var AppSetupCompleted bool = false

// PublicPath is the root directory of the static public folder.
// It is set from the "-a" flag in main.go and used by upload/thumbnail
// helpers to resolve the storage location dynamically.
var PublicPath string = "./public"

// SessStore is a global variable for session management
var SessStore *session.Store

func InitSession() {
	SessStore = session.New(session.Config{
		// Default expiration if "Remember Me" is not checked (e.g. 2 hours)
		Expiration: 2 * time.Hour, 
		// HTTP Only and SameSite cookies for basic XSS & CSRF security
		CookieHTTPOnly: true,
		CookieSameSite: "Strict",
	})
}
