package seeder

import (
	"log"

	"gorm.io/gorm"
)

// SyncThemes synchronizes the hardcoded theme registry with the database.
// This ensures that the static frontend themes are always available as foreign keys
// in the database for the Project entity.
func SyncThemes(db *gorm.DB) {
	for _, theme := range ThemesData {
		result := db.Where("id = ?", theme.ID).FirstOrCreate(&theme)
		if result.Error != nil {
			log.Printf("Failed to sync theme %s: %v", theme.ID, result.Error)
		} else {
			// Update if it already exists, but DO NOT overwrite UI-managed fields
			db.Model(&theme).Select("Name", "EventType", "MediaBuckets").Updates(theme)
		}
	}
	log.Println("Seeded themes successfully.")
}
