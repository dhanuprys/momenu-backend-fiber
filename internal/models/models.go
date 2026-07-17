package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// EventType defines the type of event for strong typing
type EventType string

const (
	EventTypePernikahan  EventType = "pernikahan"
	EventTypeUlangTahun  EventType = "ulang_tahun"
	EventTypeMetatah     EventType = "metatah"
	EventTypeTigangSasih EventType = "tigang_sasih"
	EventTypeSeminar     EventType = "seminar"
)

// GiftRegistryType defines the type of gift registry
type GiftRegistryType string

const (
	GiftRegistryTypeBank     GiftRegistryType = "bank"
	GiftRegistryTypeEWallet  GiftRegistryType = "ewallet"
	GiftRegistryTypePhysical GiftRegistryType = "physical"
)

// ProjectStatus defines the state of a project
type ProjectStatus string

const (
	ProjectStatusDraft     ProjectStatus = "draft"
	ProjectStatusPublished ProjectStatus = "published"
	ProjectStatusArchived  ProjectStatus = "archived"
)

// MediaType defines the type of media for media mappings and buckets
type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
)

// LiveStreamPlatform defines the platform for live streaming
type LiveStreamPlatform string

const (
	PlatformYouTube   LiveStreamPlatform = "youtube"
	PlatformZoom      LiveStreamPlatform = "zoom"
	PlatformInstagram LiveStreamPlatform = "instagram"
	PlatformTiktok    LiveStreamPlatform = "tiktok"
	PlatformGmeet     LiveStreamPlatform = "gmeet"
	PlatformOther     LiveStreamPlatform = "other"
)

// FieldType defines the type of form field for the field schema
type FieldType string

const (
	FieldTypeString FieldType = "string"
	FieldTypeText   FieldType = "text"   // multiline
	FieldTypeNumber FieldType = "number"
	FieldTypeURL    FieldType = "url"
	FieldTypeGroup  FieldType = "group" // repeatable array of sub-fields
	FieldTypeImage  FieldType = "image" // single image upload (stores file path)
)

// FieldDefinition describes a single form field in the event-type schema
type FieldDefinition struct {
	Key         string            `json:"key"`
	Label       string            `json:"label"`
	Type        FieldType         `json:"type"`
	Required    bool              `json:"required"`
	Placeholder string            `json:"placeholder,omitempty"`
	Validations []string          `json:"validations,omitempty"`
	Fields      []FieldDefinition `json:"fields,omitempty"`    // sub-fields for group type
	MinItems    int               `json:"min_items,omitempty"` // minimum entries for group type
	MaxItems    int               `json:"max_items,omitempty"` // maximum entries for group type
}

// FieldGroup represents a visual group of fields in the builder form
type FieldGroup struct {
	GroupName string            `json:"group_name"`
	Fields    []FieldDefinition `json:"fields"`
}

// MediaBucket describes a media upload slot tied to a theme's visual layout
type MediaBucket struct {
	Key       string    `json:"key"`
	Label     string    `json:"label"`
	MediaType MediaType `json:"media_type"` // strongly-typed media type
	MaxFiles  int       `json:"max_files"`
}

// EventTypeFieldSchemas maps each EventType to its required form definition.
// This is the single source of truth for what data each event type needs.
var EventTypeFieldSchemas = map[EventType][]FieldGroup{
	EventTypePernikahan: {
		{
			GroupName: "Informasi Mempelai Pria",
			Fields: []FieldDefinition{
				{Key: "nama_mempelai_pria", Label: "Nama Mempelai Pria", Type: FieldTypeString, Required: true, Placeholder: "Masukkan nama lengkap mempelai pria", Validations: []string{"min:2", "max:100"}},
				{Key: "nama_panggilan_pria", Label: "Nama Panggilan Pria", Type: FieldTypeString, Required: true, Placeholder: "Masukkan nama panggilan pria", Validations: []string{"min:2", "max:50"}},
				{Key: "nama_ayah_pria", Label: "Nama Ayah Mempelai Pria", Type: FieldTypeString, Required: true, Placeholder: "Masukkan nama ayah mempelai pria", Validations: []string{"min:2", "max:100"}},
				{Key: "nama_ibu_pria", Label: "Nama Ibu Mempelai Pria", Type: FieldTypeString, Required: true, Placeholder: "Masukkan nama ibu mempelai pria", Validations: []string{"min:2", "max:100"}},
				{Key: "anak_ke_pria", Label: "Anak Ke", Type: FieldTypeNumber, Required: false, Validations: []string{"min:1", "max:20"}},
				{Key: "bersaudara_pria", Label: "Dari Berapa Bersaudara", Type: FieldTypeNumber, Required: false, Validations: []string{"min:1", "max:20"}},
				{Key: "alamat_pria", Label: "Alamat Mempelai Pria", Type: FieldTypeText, Required: false, Placeholder: "Masukkan alamat mempelai pria", Validations: []string{"max:300"}},
			},
		},
		{
			GroupName: "Informasi Mempelai Wanita",
			Fields: []FieldDefinition{
				{Key: "nama_mempelai_wanita", Label: "Nama Mempelai Wanita", Type: FieldTypeString, Required: true, Placeholder: "Masukkan nama lengkap mempelai wanita", Validations: []string{"min:2", "max:100"}},
				{Key: "nama_panggilan_wanita", Label: "Nama Panggilan Wanita", Type: FieldTypeString, Required: true, Placeholder: "Masukkan nama panggilan wanita", Validations: []string{"min:2", "max:50"}},
				{Key: "nama_ayah_wanita", Label: "Nama Ayah Mempelai Wanita", Type: FieldTypeString, Required: true, Placeholder: "Masukkan nama ayah mempelai wanita", Validations: []string{"min:2", "max:100"}},
				{Key: "nama_ibu_wanita", Label: "Nama Ibu Mempelai Wanita", Type: FieldTypeString, Required: true, Placeholder: "Masukkan nama ibu mempelai wanita", Validations: []string{"min:2", "max:100"}},
				{Key: "anak_ke_wanita", Label: "Anak Ke", Type: FieldTypeNumber, Required: false, Validations: []string{"min:1", "max:20"}},
				{Key: "bersaudara_wanita", Label: "Dari Berapa Bersaudara", Type: FieldTypeNumber, Required: false, Validations: []string{"min:1", "max:20"}},
				{Key: "alamat_wanita", Label: "Alamat Mempelai Wanita", Type: FieldTypeText, Required: false, Placeholder: "Masukkan alamat mempelai wanita", Validations: []string{"max:300"}},
			},
		},
	},
	EventTypeUlangTahun: {
		{
			GroupName: "Informasi",
			Fields: []FieldDefinition{
				{Key: "nama_dirayakan", Label: "Nama yang Dirayakan", Type: FieldTypeString, Required: true, Placeholder: "Masukkan nama lengkap", Validations: []string{"min:2", "max:100"}},
				{Key: "umur", Label: "Umur Saat Ini", Type: FieldTypeNumber, Required: true, Validations: []string{"min:1", "max:150"}},
			},
		},
	},
	EventTypeMetatah: {
		{
			GroupName: "Peserta Metatah",
			Fields: []FieldDefinition{
				{
					Key: "peserta", Label: "Peserta", Type: FieldTypeGroup, Required: true,
					MinItems: 1, MaxItems: 5,
					Fields: []FieldDefinition{
						{Key: "nama", Label: "Nama Peserta", Type: FieldTypeString, Required: true, Placeholder: "Masukkan nama peserta", Validations: []string{"min:2", "max:100"}},
						{Key: "foto", Label: "Foto Peserta", Type: FieldTypeImage, Required: false},
					},
				},
			},
		},
		{
			GroupName: "Informasi Orang Tua",
			Fields: []FieldDefinition{
				{Key: "nama_ayah", Label: "Nama Ayah", Type: FieldTypeString, Required: true, Placeholder: "Masukkan nama ayah", Validations: []string{"min:2", "max:100"}},
				{Key: "nama_ibu", Label: "Nama Ibu", Type: FieldTypeString, Required: true, Placeholder: "Masukkan nama ibu", Validations: []string{"min:2", "max:100"}},
			},
		},
	},
	EventTypeTigangSasih: {
		{
			GroupName: "Informasi Bayi",
			Fields: []FieldDefinition{
				{Key: "nama_bayi", Label: "Nama Bayi", Type: FieldTypeString, Required: true, Placeholder: "Masukkan nama bayi", Validations: []string{"min:2", "max:100"}},
			},
		},
		{
			GroupName: "Informasi Orang Tua",
			Fields: []FieldDefinition{
				{Key: "nama_ayah", Label: "Nama Ayah", Type: FieldTypeString, Required: true, Placeholder: "Masukkan nama ayah", Validations: []string{"min:2", "max:100"}},
				{Key: "nama_ibu", Label: "Nama Ibu", Type: FieldTypeString, Required: true, Placeholder: "Masukkan nama ibu", Validations: []string{"min:2", "max:100"}},
			},
		},
	},
	EventTypeSeminar: {
		{
			GroupName: "Informasi Seminar",
			Fields: []FieldDefinition{
				{Key: "nama_pembicara", Label: "Nama Pembicara", Type: FieldTypeString, Required: true, Placeholder: "Masukkan nama pembicara", Validations: []string{"min:2", "max:100"}},
				{Key: "nama_seminar", Label: "Nama Seminar", Type: FieldTypeString, Required: true, Placeholder: "Masukkan nama/judul seminar", Validations: []string{"min:2", "max:200"}},
			},
		},
	},
}

// User - Identity & Authority
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email"`
	Password  string         `json:"-"` // Not exposed in JSON
	GoogleID  *string        `gorm:"uniqueIndex" json:"google_id,omitempty"`
	IsAdmin   bool           `gorm:"default:false" json:"is_admin"`
	Verified  bool           `gorm:"default:false" json:"verified"`
	Projects  []Project      `gorm:"foreignKey:UserID" json:"projects,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Theme - Visual Registry
type Theme struct {
	ID          string    `gorm:"primaryKey" json:"id"`       // e.g., "floral_wedding"
	Name        string    `gorm:"not null" json:"name"`       // e.g., "Floral Elegance"
	EventType   EventType `gorm:"not null;index" json:"event_type"` // strongly-typed event type
	Description string    `json:"description"`
	Thumbnail    string         `json:"thumbnail"`
	Price        *float64       `json:"price"`         // nullable price
	MediaBuckets datatypes.JSON `json:"media_buckets"` // []MediaBucket — theme-specific gallery sections
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	UsageCount   int64          `gorm:"-" json:"usage_count,omitempty"`
}

// MusicCategory - Curated music categories matching EventType
type MusicCategory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`             // e.g., "Wedding", "Birthday"
	Slug        string    `gorm:"uniqueIndex;not null" json:"slug"` // matches EventType
	Description string    `json:"description"`
	Order       int       `gorm:"default:0" json:"order"`
}

// Music - Curated audio tracks for projects
type Music struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	CategoryID      uint          `gorm:"not null" json:"category_id"`
	Category        MusicCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Title           string        `gorm:"not null" json:"title"`
	Artist          string        `gorm:"not null" json:"artist"`
	FilePath        string        `gorm:"not null" json:"file_path"`
	DurationSeconds int           `json:"duration_seconds"`
	CoverImage      string        `json:"cover_image"`
	Order           int           `gorm:"default:0" json:"order"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// Project - Core Project Architecture
type Project struct {
	ID               uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID           uint           `gorm:"not null;index" json:"user_id"`
	Title            string         `gorm:"not null" json:"title"`
	ThemeID          string         `gorm:"not null" json:"theme_id"` // Matches hardcoded theme registry
	Theme            Theme          `gorm:"foreignKey:ThemeID" json:"theme,omitempty"`
	MusicID          *uint          `json:"music_id"`
	Music            Music          `gorm:"foreignKey:MusicID;constraint:OnDelete:SET NULL;" json:"music,omitempty"`
	EventType        EventType      `gorm:"not null" json:"event_type"` // strongly-typed event type
	Status           ProjectStatus  `gorm:"default:'draft'" json:"status"`
	Slug             string         `gorm:"uniqueIndex;not null" json:"slug"`
	SharingThumbnail string         `json:"sharing_thumbnail"`
	FeatureToggle    FeatureToggle  `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"feature_toggle"`
	Payload          datatypes.JSON `json:"payload"`
	Schedules        []Schedule     `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"schedules"`
	GiftRegistries   []GiftRegistry `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"gift_registries"`
	MediaMappings    []MediaMapping `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"media_mappings"`
	DressCodes       []DressCode     `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"dress_codes"`
	LiveStreams      []LiveStream    `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"live_streams"`
	RSVPs            []RSVP          `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"rsvps"`
	Guestbooks       []Guestbook     `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"guestbooks"`
	TextOverrides    []TextOverride  `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"text_overrides"`
	StyleOverrides   []StyleOverride `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"style_overrides"`
	ProjectVisits    []ProjectVisit  `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"project_visits"`
	ShareSessions    []ProjectShareSession `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"share_sessions"`
	DiskQuotaBytes   int64          `gorm:"default:104857600" json:"disk_quota_bytes"` // default 100MB
	UpdateCount      uint           `gorm:"default:0" json:"update_count"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// FeatureToggle - UI Switchboard
type FeatureToggle struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ProjectID      uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"project_id"`
	ShowRSVP       bool `gorm:"default:true" json:"show_rsvp"`
	ShowWishes             bool      `gorm:"default:true" json:"show_wishes"`
	ShowGallery            bool      `gorm:"default:true" json:"show_gallery"`
	ShowGifts              bool      `gorm:"default:true" json:"show_gifts"`
	ShowLiveStream         bool      `gorm:"default:false" json:"show_live_stream"`
	ShowMusic              bool      `gorm:"default:true" json:"show_music"`
	RequireRegisteredGuest bool      `gorm:"default:false" json:"require_registered_guest"`
	WhatsappTemplate       string    `gorm:"type:text" json:"whatsapp_template"`
}

// GetFieldSchema returns the field schema for a given EventType.
// Returns nil if the event type is not registered.
func GetFieldSchema(eventType EventType) []FieldGroup {
	schema, exists := EventTypeFieldSchemas[eventType]
	if !exists {
		return nil
	}
	return schema
}

// Schedule - Event Content & Timeline
type Schedule struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
	Title     string    `gorm:"not null" json:"title"` // e.g., "Morning Ceremony"
	StartTime time.Time `gorm:"not null" json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Timezone  string    `gorm:"not null" json:"timezone"`
	Location  string         `json:"location"`
	MapURL    string         `json:"map_url"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// GiftRegistry - Financial/Gifting Router
type GiftRegistry struct {
	ID             uint             `gorm:"primaryKey" json:"id"`
	ProjectID      uuid.UUID        `gorm:"type:uuid;not null;index" json:"project_id"`
	Type           GiftRegistryType `gorm:"not null" json:"type"` // strongly-typed gift registry type
	ProviderName   string           `json:"provider_name"`        // e.g., "BCA", "GoPay"
	AccountNumber  string           `json:"account_number"`
	AccountName    string           `json:"account_name"`
	QRCodeImage    string           `json:"qr_code_image"`
	PhoneNumber    string           `json:"phone_number"` // For e-wallet
	MailingAddress string           `json:"mailing_address"`
	DeletedAt      gorm.DeletedAt   `gorm:"index" json:"-"`
}

// MediaMapping - Galleries & Teasers
type MediaMapping struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index:idx_media_project_bucket,priority:1;index" json:"project_id"`
	Bucket    string    `gorm:"not null;index:idx_media_project_bucket,priority:2" json:"bucket"`     // e.g., "hero_slider", "second_section_grid", "footer_teaser"
	MediaType MediaType `gorm:"not null" json:"media_type"` // strongly-typed media type
	URL       string         `gorm:"not null" json:"url"`
	Order     int            `gorm:"default:0" json:"order"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// RSVP - Guest Interactions
type RSVP struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ProjectID      uuid.UUID      `gorm:"type:uuid;uniqueIndex:idx_rsvp_project_name;index:idx_rsvp_stats,priority:1;not null" json:"project_id"`
	Name           string         `gorm:"uniqueIndex:idx_rsvp_project_name;not null" json:"name"`
	Attending      bool           `gorm:"not null;index:idx_rsvp_stats,priority:2" json:"attending"`
	GuestCount     int            `gorm:"default:1" json:"guest_count"`
	SpecialMessage string         `gorm:"type:text" json:"special_message,omitempty"`
	Whatsapp       *string        `json:"whatsapp,omitempty"`
	IsResponded    bool           `gorm:"default:false;index:idx_rsvp_stats,priority:3" json:"is_responded"`
	HasOpened      bool           `gorm:"default:false" json:"has_opened"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// Guestbook - Wishes
type Guestbook struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
	Name      string    `gorm:"not null" json:"name"`
	Message   string         `gorm:"type:text;not null" json:"message"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TextOverride - per-project text content customization
type TextOverride struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_text_override_slot" json:"project_id"`
	SlotKey   string    `gorm:"not null;uniqueIndex:idx_text_override_slot" json:"slot_key"`
	Value     string    `gorm:"type:text;not null" json:"value"`
	Bold      bool      `gorm:"default:false" json:"bold"`
	Italic    bool      `gorm:"default:false" json:"italic"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StyleOverride - per-project container/visual customization
type StyleOverride struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	ProjectID  uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_style_override_slot" json:"project_id"`
	SlotKey    string         `gorm:"not null;uniqueIndex:idx_style_override_slot" json:"slot_key"`
	Properties datatypes.JSON `gorm:"type:jsonb;not null" json:"properties"` // {"backgroundColor":"#fff",...}
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// RSVPStats - Non-DB struct for aggregate data

// DressCode - Flexible Dress Code Router
type DressCode struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ProjectID uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	Label     string         `gorm:"not null" json:"label"`  // e.g., "Male/Female", "L/P"
	Colors    datatypes.JSON `gorm:"not null" json:"colors"` // JSON list of colors
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// LiveStream - Dedicated resource for live streaming
type LiveStream struct {
	ID        uint               `gorm:"primaryKey" json:"id"`
	ProjectID uuid.UUID          `gorm:"type:uuid;not null;index" json:"project_id"`
	Platform  LiveStreamPlatform `gorm:"not null" json:"platform"` // strongly-typed platform
	URL       string             `gorm:"not null" json:"url"`
	CreatedAt time.Time          `json:"created_at"`
	DeletedAt gorm.DeletedAt     `gorm:"index" json:"-"`
}

// ProjectVisit - Analytics Tracking
type ProjectVisit struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ProjectID uuid.UUID      `gorm:"type:uuid;index;index:idx_visit_project_source,priority:1;index:idx_visit_project_device,priority:1;index:idx_visit_project_ip,priority:1;index:idx_visit_project_created,priority:1;not null" json:"project_id"`
	GuestName string         `json:"guest_name"`
	Source    string         `gorm:"index:idx_visit_project_source,priority:2" json:"source"`
	UserAgent  string         `json:"user_agent"`
	DeviceType string         `gorm:"index:idx_visit_project_device,priority:2" json:"device_type"`
	IPAddress  string         `gorm:"index:idx_visit_project_ip,priority:2" json:"ip_address"`
	Country    *string        `json:"country,omitempty"`
	CreatedAt time.Time      `gorm:"index:idx_visit_project_created,priority:2" json:"created_at"`
}

// ProjectShareSession - Tracks active share sessions
type ProjectShareSession struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ProjectID uuid.UUID      `gorm:"type:uuid;index;not null" json:"project_id"`
	Name      string         `json:"name"`
	SessionID string         `gorm:"uniqueIndex;not null" json:"session_id"`
	IsRevoked bool           `gorm:"default:false" json:"is_revoked"`
	ExpiresAt *time.Time     `json:"expires_at"`
	LastAccessedAt *time.Time `json:"last_accessed_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// FileRecord represents a file uploaded by a user and stored on disk.
type FileRecord struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	URL           string         `gorm:"uniqueIndex;not null" json:"url"`
	FilePath      string         `gorm:"not null" json:"file_path"`
	OriginalName  string         `json:"original_name"`
	ContentType   string         `json:"content_type"`
	Size          int64          `gorm:"not null" json:"size"`
	OptimizedSize *int64         `json:"optimized_size,omitempty"`
	IsOptimized   bool           `gorm:"default:false;index:idx_file_unoptimized,priority:1" json:"is_optimized"`
	MediaType     string         `gorm:"not null;index" json:"media_type"`
	ProjectID     *uuid.UUID     `gorm:"type:uuid;index" json:"project_id,omitempty"`
	UploadedByID  *uint          `gorm:"index" json:"uploaded_by_id,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
