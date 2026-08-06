package models

import (
	"gorm.io/gorm"
)

// Helper to safely get setting value
func GetSetting(db *gorm.DB, key string, defaultValue string) string {
	var setting Setting
	// Since Key is primary key, we search directly in the key column
	result := db.Where("key = ?", key).First(&setting)
	if result.Error != nil {
		return defaultValue // Return default value if key is not set in database
	}
	return setting.Value
}
