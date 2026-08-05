package database

import (
	"log"
	"go-lightblog/config"
	"go-lightblog/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB adalah instance global agar bisa dipanggil dari handlers/services
var DB *gorm.DB

func Connect(dbPath string) {
	// Menggunakan glebarez/sqlite pure Go
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent agar terminal rapi, bisa diganti ke Info jika ingin melihat raw SQL
	})
	if err != nil {
		log.Fatalf("Gagal terkoneksi ke database SQLite: %v", err)
	}

	// Menjalankan migrasi otomatis untuk struktur tabel
	err = db.AutoMigrate(
		&models.User{}, 
		&models.Setting{}, 
		&models.Category{}, 
		&models.Tag{}, 
		&models.Post{},
	)
	if err != nil {
		log.Fatalf("Gagal melakukan migrasi database: %v", err)
	}

	// ==========================================
	// SEEDER PENGATURAN DEFAULT CMS
	// ==========================================
	defaultSettings := []models.Setting{
		{Key: "site_description", Value: "Blog minimalis dan cepat bertenaga Go Fiber."},
		{Key: "site_keywords", Value: "blog, go, fiber, lightblog"},
		{Key: "upload_mode", Value: "local"},
		{Key: "imagekit_private_key", Value: ""},
		{Key: "imagekit_folder", Value: "/lightblog"},
		{Key: "remark42_url", Value: "http://127.0.0.1:8080"},
		{Key: "remark42_site_id", Value: "lightblog-utama"},
		{Key: "enable_gemini", Value: "no"},
		{Key: "gemini_api_key", Value: ""},
		{Key: "gemini_model", Value: "gemini-flash-latest"},
	}

	// Looping untuk memastikan setiap konfigurasi dasar memiliki baris di database
	for _, setting := range defaultSettings {
		// FirstOrCreate akan mencari data berdasarkan Key. 
		// Jika tidak ada, ia akan melakukan INSERT. Jika sudah ada, ia akan diam (mengabaikan).
		// Ini memastikan pengaturan yang sudah diubah oleh admin tidak tertimpa ulang saat server restart.
		db.FirstOrCreate(&setting, models.Setting{Key: setting.Key})
	}
	// ==========================================

	// Mengecek apakah sudah ada data admin di tabel User
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount > 0 {
		config.AppSetupCompleted = true
		log.Println("Status CMS: Setup sudah selesai.")
	} else {
		config.AppSetupCompleted = false
		log.Println("Status CMS: Membutuhkan inisialisasi awal (Setup Mode).")
	}

	log.Println("Database SQLite berhasil terhubung dan dimigrasi.")
	DB = db
}