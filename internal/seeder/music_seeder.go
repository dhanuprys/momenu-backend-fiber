package seeder

import (
	"log"

	"gorm.io/gorm"
)

// SyncMusicCategories synchronizes the hardcoded music categories with the database.
func SyncMusicCategories(db *gorm.DB) {
	for _, category := range MusicCategoriesData {
		result := db.Where("id = ?", category.ID).FirstOrCreate(&category)
		if result.Error != nil {
			log.Printf("Failed to sync music category %s: %v", category.Slug, result.Error)
		} else {
			// Update if it already exists to ensure fields are fresh
			db.Model(&category).Updates(category)
		}
	}
	log.Println("Seeded music categories successfully.")
}

// SyncMusics synchronizes the dummy music tracks with the database.
func SyncMusics(db *gorm.DB) {
	for _, music := range MusicsData {
		result := db.Where("id = ?", music.ID).FirstOrCreate(&music)
		if result.Error != nil {
			log.Printf("Failed to sync music %s: %v", music.Title, result.Error)
		} else {
			// Update if it already exists to ensure fields are fresh
			db.Model(&music).Updates(music)
		}
	}
	log.Println("Seeded musics successfully.")
}
