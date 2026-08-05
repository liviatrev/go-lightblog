package models

import (
	"gorm.io/gorm"
)

// Helper untuk mengambil nilai pengaturan dengan aman
func GetSetting(db *gorm.DB, key string, defaultValue string) string {
	var setting Setting
	// Karena Key adalah primary key, kita mencarinya langsung di kolom key
	result := db.Where("key = ?", key).First(&setting)
	if result.Error != nil {
		return defaultValue // Mengembalikan nilai default jika key belum diatur di database
	}
	return setting.Value
}