package seeder

import (
	"encoding/json"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"gorm.io/datatypes"
)

func pricePtr(p float64) *float64 {
	return &p
}

func mustJSON(v interface{}) datatypes.JSON {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

var ThemesData = []models.Theme{
	{
		ID:          "pernikahan_bali_simple_1",
		Name:        "Pernikahan Bali Simple",
		EventType:   models.EventTypePernikahan,
		Description: "Tema pawiwahan bernuansa adat Bali yang sederhana, elegan, dan informatif.",
		Thumbnail:   "https://images.unsplash.com/photo-1544928147-79a2dbc1f389?q=80&w=1974&auto=format&fit=crop",
		Price:       pricePtr(79000),
		MediaBuckets: mustJSON([]models.MediaBucket{
			{Key: "hero_photo", Label: "Foto Sampul", MediaType: models.MediaTypeImage, MaxFiles: 1},
			{Key: "groom_photo", Label: "Foto Mempelai Pria", MediaType: models.MediaTypeImage, MaxFiles: 1},
			{Key: "bride_photo", Label: "Foto Mempelai Wanita", MediaType: models.MediaTypeImage, MaxFiles: 1},
			{Key: "gallery_grid", Label: "Galeri Pre-Wedding", MediaType: models.MediaTypeImage, MaxFiles: 20},
			{Key: "promo_video", Label: "Video Cerita Kami", MediaType: models.MediaTypeVideo, MaxFiles: 1},
		}),
	},
	{
		ID:          "ulang_tahun_festive_1",
		Name:        "Festive Party",
		EventType:   models.EventTypeUlangTahun,
		Description: "Tema pesta ulang tahun yang meriah dan penuh warna.",
		Thumbnail:   "https://images.unsplash.com/photo-1555243896-c709bfa0b564?q=80&w=1770&auto=format&fit=crop",
		Price:       pricePtr(29.99),
		MediaBuckets: mustJSON([]models.MediaBucket{
			{Key: "cover_photo", Label: "Foto Sampul", MediaType: models.MediaTypeImage, MaxFiles: 1},
			{Key: "gallery_grid", Label: "Galeri Foto", MediaType: models.MediaTypeImage, MaxFiles: 10},
		}),
	},
	{
		ID:          "metatah_sakral_1",
		Name:        "Sakral Bali",
		EventType:   models.EventTypeMetatah,
		Description: "Tema sakral untuk upacara potong gigi khas Bali yang khidmat.",
		Thumbnail:   "https://images.unsplash.com/photo-1604999333679-b86d54738315?q=80&w=2025&auto=format&fit=crop",
		Price:       pricePtr(39.99),
		MediaBuckets: mustJSON([]models.MediaBucket{
			{Key: "cover_photo", Label: "Foto Sampul", MediaType: models.MediaTypeImage, MaxFiles: 1},
			{Key: "ceremony_gallery", Label: "Galeri Upacara", MediaType: models.MediaTypeImage, MaxFiles: 15},
		}),
	},
	{
		ID:          "tigang_sasih_pastel_1",
		Name:        "Pastel Joy",
		EventType:   models.EventTypeTigangSasih,
		Description: "Tema pastel lembut untuk upacara tiga bulanan bayi yang penuh kebahagiaan.",
		Thumbnail:   "https://images.unsplash.com/photo-1519689680058-324335c77eba?q=80&w=2070&auto=format&fit=crop",
		Price:       pricePtr(19.99),
		MediaBuckets: mustJSON([]models.MediaBucket{
			{Key: "cover_photo", Label: "Foto Sampul", MediaType: models.MediaTypeImage, MaxFiles: 1},
			{Key: "baby_gallery", Label: "Galeri Bayi", MediaType: models.MediaTypeImage, MaxFiles: 10},
		}),
	},
	{
		ID:          "seminar_professional_1",
		Name:        "Professional",
		EventType:   models.EventTypeSeminar,
		Description: "Tampilan profesional dan modern untuk seminar dan acara formal.",
		Thumbnail:   "https://images.unsplash.com/photo-1540317580384-e5d43616b9aa?q=80&w=1974&auto=format&fit=crop",
		Price:       pricePtr(99.99),
		MediaBuckets: mustJSON([]models.MediaBucket{
			{Key: "hero_banner", Label: "Banner Utama", MediaType: models.MediaTypeImage, MaxFiles: 1},
			{Key: "speaker_photos", Label: "Foto Pembicara", MediaType: models.MediaTypeImage, MaxFiles: 5},
		}),
	},
}
