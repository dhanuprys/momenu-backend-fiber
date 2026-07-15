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

var MusicsData = []models.Music{}
