package seeder

import (
	"log"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"gorm.io/gorm"
)

// SyncThemes synchronizes the hardcoded theme registry with the database.
// This ensures that the static frontend themes are always available as foreign keys
// in the database for the Project entity.
func SyncThemes(db *gorm.DB) {
	for _, theme := range ThemesData {
		var existing models.Theme
		result := db.Where("id = ?", theme.ID).First(&existing)
		if result.Error != nil {
			if err := db.Create(&theme).Error; err != nil {
				log.Printf("Failed to create theme %s: %v", theme.ID, err)
			}
		} else {
			// Update if it already exists, but DO NOT overwrite UI-managed fields
			db.Model(&existing).Select("Name", "EventType", "MediaBuckets").Updates(theme)
		}
	}
	log.Println("Seeded themes successfully.")
}
