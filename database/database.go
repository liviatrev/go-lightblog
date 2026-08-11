package database

import (
	"log"
	"time"

	"go-lightblog/config"
	"go-lightblog/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is a global instance so it can be called from handlers/services
var DB *gorm.DB

func Connect(dbPath string) {
	// 1. Using glebarez/sqlite pure Go and PRAGMA params
	dsn := dbPath + "?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent to keep terminal clean, can change to Info to see raw SQL
	})
	if err != nil {
		log.Fatalf("Failed to connect to SQLite database: %v", err)
	}

	// 2. Prevent Memory Leaks
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to extract sql.DB to set connection pool: %v", err)
	}
	
	// Limit connection
	sqlDB.SetMaxIdleConns(5)                  
	sqlDB.SetMaxOpenConns(25)                 
	sqlDB.SetConnMaxLifetime(1 * time.Hour)

	// Run automatic migration for table structure
	err = db.AutoMigrate(
		&models.User{}, 
		&models.Setting{}, 
		&models.Category{}, 
		&models.Tag{}, 
		&models.Post{},
		&models.SlugRedirect{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// ==========================================
	// DEFAULT CMS SETTINGS SEEDER
	// ==========================================
	defaultSettings := []models.Setting{
		{Key: "site_description", Value: "A minimal and fast blog powered by Go Fiber."},
		{Key: "site_keywords", Value: "blog, go, fiber, lightblog"},
		{Key: "site_headline", Value: "Explore Articles"},
		{Key: "site_tagline", Value: "A collection of the latest writings, notes, and insights."},
		{Key: "upload_mode", Value: "local"},
		{Key: "imagekit_private_key", Value: ""},
		{Key: "imagekit_folder", Value: "/lightblog"},
		{Key: "remark42_url", Value: "http://127.0.0.1:8080"},
		{Key: "remark42_site_id", Value: "lightblog-utama"},
		{Key: "enable_gemini", Value: "no"},
		{Key: "gemini_api_key", Value: ""},
		{Key: "gemini_model", Value: "gemini-flash-latest"},
		{Key: "enable_cloudflare", Value: "no"},
		{Key: "cloudflare_api_key", Value: ""},
		{Key: "cloudflare_zone_id", Value: ""},
		{Key: "site_url", Value: ""},
		{Key: "public_theme", Value: "light"},
	}

	// Loop to ensure every basic configuration has a row in the database
	for _, setting := range defaultSettings {
		// FirstOrCreate will search data by Key. 
		// If not found, it will INSERT. If found, it will stay silent (ignore).
		// This ensures settings already changed by admin are not overwritten on server restart.
		db.FirstOrCreate(&setting, models.Setting{Key: setting.Key})
	}
	// ==========================================

	// Check if there is already admin data in User table
	var userCount int64
	if err := db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		log.Fatalf("Failed to query user count during setup check: %v", err)
	}
	if userCount > 0 {
		config.AppSetupCompleted = true
		log.Println("CMS Status: Setup completed.")
	} else {
		config.AppSetupCompleted = false
		log.Println("CMS Status: Requires initial setup (Setup Mode).")
	}

	log.Println("SQLite database connected and migrated successfully.")
	DB = db
}
