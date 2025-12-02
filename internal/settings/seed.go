package settings

// This whole thing is a little bit fragile and might change a bit more.
// I don't like how everything is a string that has to be inferred lmao
// Surely there's a better way to do this?

// I guess this is just the first pass of implementing the default/overide pattern
// https://web.archive.org/web/20250706041703/https://double.finance/blog/default_override

import (
	"log/slog"

	"gorm.io/gorm"
	
	"imagine/internal/entities"
)

var defaultSettings = []entities.SettingDefault{
	{
		Name:           "theme",
		Value:          "system",
		ValueType:      "enum",
		AllowedValues:  &[]string{"Light", "Dark", "System"},
		IsUserEditable: true,
		Group:          "General",
		Description:    "Choose your preferred theme: light, dark, or system default.",
	},
	{
		Name:           "language",
		Value:          "en-GB",
		ValueType:      "string",
		IsUserEditable: true,
		Group:          "General",
		Description:    "Your preferred display language (ISO 639-1 code, e.g., en-ZA).",
	},
	{
		Name:           "timezone",
		Value:          "Africa/Johannesburg",
		ValueType:      "string",
		IsUserEditable: true,
		Group:          "General",
		Description:    "Your current timezone (IANA database identifier, e.g., Africa/Johannesburg).",
	},
	{
		Name:           "notifications_email",
		Value:          "true",
		ValueType:      "boolean",
		IsUserEditable: true,
		Group:          "Notifications",
		Description:    "Receive email notifications. (Not Implemented Yet)",
	},
	{
		Name:           "notifications_push",
		Value:          "false",
		ValueType:      "boolean",
		IsUserEditable: true,
		Group:          "Notifications",
		Description:    "Receive push notifications (Not Implemented Yet)",
	},
	{
		Name:           "privacy_profile_visibility",
		Value:          "private",
		ValueType:      "enum",
		AllowedValues:  &[]string{"Public", "Private"},
		IsUserEditable: true,
		Group:          "Privacy",
		Description:    "Control who can see your profile.",
	},
	{
		Name:           "ui_page_size_images",
		Value:          "100",
		ValueType:      "integer",
		AllowedValues:  &[]string{"50", "100", "250", "500"},
		IsUserEditable: true,
		Group:          "Interface",
		Description:    "Number of images to display per page in galleries.",
	},
	{
		Name:           "ui_page_size_collections",
		Value:          "20",
		ValueType:      "integer",
		AllowedValues:  &[]string{"20", "50", "100"},
		IsUserEditable: true,
		Group:          "Interface",
		Description:    "Number of collections to display per page.",
	},
	{
		Name:           "ui_default_view_mode",
		Value:          "masonry",
		ValueType:      "enum",
		AllowedValues:  &[]string{"grid", "masonry", "list", "feed"},
		IsUserEditable: true,
		Group:          "Interface",
		Description:    "Default display mode for image galleries.",
	},
	{
		Name:           "image_download_quality",
		Value:          "90",
		ValueType:      "integer",
		IsUserEditable: true,
		Group:          "Images",
		Description:    "Default quality (1-100) for downloaded images when format conversion occurs.",
	},
	{
		Name:           "image_download_format",
		Value:          "original",
		ValueType:      "enum",
		AllowedValues:  &[]string{"original", "jpg", "png", "webp", "avif"},
		IsUserEditable: true,
		Group:          "Images",
		Description:    "Default file format for downloaded images.",
	},
	{
		Name:           "image_preview_format",
		Value:          "webp",
		ValueType:      "enum",
		AllowedValues:  &[]string{"webp", "avif", "jpg", "png"},
		IsUserEditable: true,
		Group:          "Images",
		Description:    "Preferred file format for image previews in the browser.",
	},
	{
		Name:           "image_resize_kernel",
		Value:          "lanczos3",
		ValueType:      "enum",
		AllowedValues:  &[]string{"nearest", "linear", "cubic", "mitchell", "lanczos2", "lanczos3", "mks2013", "mks2021"},
		IsUserEditable: true,
		Group:          "Images",
		Description:    "Resampling kernel used for image resizing (e.g., 'lanczos3' for photos, 'nearest' for pixel art).",
	},
	{
		Name:           "privacy_download_strip_metadata",
		Value:          "false",
		ValueType:      "boolean",
		IsUserEditable: true,
		Group:          "Privacy",
		Description:    "Automatically remove EXIF/GPS metadata when creating download links.",
	},
	{
		Name:           "image_visible_metadata",
		Value:          `["date", "camera", "iso", "aperture"]`,
		ValueType:      "json",
		IsUserEditable: true,
		Group:          "Images",
		Description:    "A JSON array of EXIF/image metadata fields to display in image detail views.",
	},
}

// SeedDefaultSettings inserts initial default settings into the database if they don't already exist.
func SeedDefaultSettings(db *gorm.DB, logger *slog.Logger) {
	for _, setting := range defaultSettings {
		var existing entities.SettingDefault
		if err := db.Where("name = ?", setting.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if createErr := db.Create(&setting).Error; createErr != nil {
					logger.Error("failed to create default setting", slog.String("setting_name", setting.Name), slog.Any("error", createErr))
				} else {
					logger.Info("created default setting", slog.String("setting_name", setting.Name))
				}
			} else {
				logger.Error("failed to query default setting", slog.String("setting_name", setting.Name), slog.Any("error", err))
			}
		}
	}
}
