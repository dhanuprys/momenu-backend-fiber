package seeder

import (
	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
)

var MusicCategoriesData = []models.MusicCategory{
	{
		ID:          1,
		Name:        "Pernikahan",
		Slug:        string(models.EventTypePernikahan),
		Description: "Musik romantis dan elegan untuk pernikahan.",
		Order:       1,
	},
	{
		ID:          2,
		Name:        "Ulang Tahun",
		Slug:        string(models.EventTypeUlangTahun),
		Description: "Musik seru dan ceria untuk pesta ulang tahun.",
		Order:       2,
	},
	{
		ID:          3,
		Name:        "Metatah",
		Slug:        string(models.EventTypeMetatah),
		Description: "Musik sakral dan khidmat untuk upacara metatah.",
		Order:       3,
	},
	{
		ID:          4,
		Name:        "Tigang Sasih",
		Slug:        string(models.EventTypeTigangSasih),
		Description: "Musik lembut dan hangat untuk upacara tiga bulanan.",
		Order:       4,
	},
	{
		ID:          5,
		Name:        "Seminar",
		Slug:        string(models.EventTypeSeminar),
		Description: "Musik profesional dan ambient untuk seminar.",
		Order:       5,
	},
}

var MusicsData = []models.Music{
	{
		ID:              1,
		CategoryID:      1, // Wedding
		Title:           "Beautiful In White (Cover)",
		Artist:          "Shane Filan",
		FilePath:        "/static/music/wedding_beautiful_in_white.mp3",
		DurationSeconds: 233,
		CoverImage:      "https://images.unsplash.com/photo-1519225421980-715cb0215aed?q=80&w=2070&auto=format&fit=crop",
		Order:           1,
	},
	{
		ID:              2,
		CategoryID:      1, // Wedding
		Title:           "A Thousand Years",
		Artist:          "Christina Perri",
		FilePath:        "/static/music/wedding_a_thousand_years.mp3",
		DurationSeconds: 285,
		CoverImage:      "https://images.unsplash.com/photo-1511285560929-80b456fea0bc?q=80&w=2069&auto=format&fit=crop",
		Order:           2,
	},
	{
		ID:              3,
		CategoryID:      2, // Birthday
		Title:           "Happy Birthday To You",
		Artist:          "Traditional",
		FilePath:        "/static/music/birthday_happy.mp3",
		DurationSeconds: 60,
		CoverImage:      "https://images.unsplash.com/photo-1530103862676-de8892bf309c?q=80&w=1770&auto=format&fit=crop",
		Order:           1,
	},
	{
		ID:              4,
		CategoryID:      3, // Corporate
		Title:           "Inspirational Background",
		Artist:          "Audio Library",
		FilePath:        "/static/music/corporate_inspirational.mp3",
		DurationSeconds: 120,
		CoverImage:      "https://images.unsplash.com/photo-1540317580384-e5d43616b9aa?q=80&w=1974&auto=format&fit=crop",
		Order:           1,
	},
	{
		ID:              5,
		CategoryID:      4, // Baby Shower
		Title:           "Twinkle Twinkle Little Star",
		Artist:          "Lullaby",
		FilePath:        "/static/music/baby_shower_twinkle.mp3",
		DurationSeconds: 150,
		CoverImage:      "https://images.unsplash.com/photo-1519689680058-324335c77eba?q=80&w=2070&auto=format&fit=crop",
		Order:           1,
	},
}
