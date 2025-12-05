package settings

// This whole thing is a little bit fragile and might change a bit more.
// I don't like how everything is a string that has to be inferred lmao
// Surely there's a better way to do this?

// I guess this is just the first pass of implementing the default/overide pattern
// https://web.archive.org/web/20250706041703/https://double.finance/blog/default_override

import (
	"errors"
	"log/slog"

	"gorm.io/gorm"

	"imagine/internal/entities"
	imaTime "imagine/internal/time"
	"imagine/internal/utils"
)

var defaultSettings = []entities.SettingDefault{
	EnumSetting(
		"theme",
		"System",
		[]string{"Light", "Dark", "System"},
		true,
		"General",
		"Choose your preferred theme: Light, Dark, or System Default.",
	),
	StringSetting(
		"language",
		"en-GB",
		true,
		"General",
		"Your preferred display language (ISO 639-1 code, e.g. en-ZA).",
	),
	EnumSetting(
		"timezone",
		"Africa/Johannesburg",
		imaTime.Timezones,
		true,
		"General",
		"Your current timezone (IANA database identifier, e.g. Africa/Johannesburg).",
	),
	BoolSetting(
		"notifications_email",
		true,
		true,
		"Notifications",
		"Receive email notifications. (Not Implemented Yet)",
	),
	BoolSetting(
		"notifications_push",
		false,
		true,
		"Notifications",
		"Receive push notifications (Not Implemented Yet)",
	),
	EnumSetting(
		"privacy_profile_visibility",
		"Private",
		[]string{"Public", "Private"},
		true,
		"Privacy",
		"Control who can see your profile.",
	),
	IntSetting(
		"ui_page_size_images",
		100,
		[]int{50, 100, 250, 500},
		true,
		"Interface",
		"Number of images to display per page in galleries.",
	),
	IntSetting(
		"ui_page_size_collections",
		20,
		[]int{20, 50, 100},
		true,
		"Interface",
		"Number of collections to display per page.",
	),
	EnumSetting(
		"ui_default_view_mode",
		"Grid",
		[]string{"Grid", "List", "Cards"},
		true,
		"Interface",
		"Default display mode for image galleries.",
	),
	IntSetting(
		"image_download_quality",
		90,
		nil,
		true,
		"Images",
		"Default quality (1-100) for downloaded images when format conversion occurs.",
	),
	EnumSetting(
		"image_download_format",
		"original",
		[]string{"original", "jpg", "png", "webp", "avif"},
		true,
		"Images",
		"Default file format for downloaded images.",
	),
	EnumSetting(
		"image_preview_format",
		"webp",
		[]string{"webp", "avif", "jpg", "png"},
		true,
		"Images",
		"Preferred file format for image previews in the browser.",
	),
	EnumSetting(
		"image_resize_kernel",
		"lanczos3",
		[]string{"nearest", "linear", "cubic", "mitchell", "lanczos2", "lanczos3", "mks2013", "mks2021"},
		true,
		"Images",
		"Resampling kernel used for image resizing (e.g., 'lanczos3' for photos, 'nearest' for pixel art).",
	),
	BoolSetting(
		"privacy_download_strip_metadata",
		false,
		true,
		"Privacy",
		"Automatically remove EXIF/GPS metadata when creating download links.",
	),
	JsonSetting(
		"image_visible_metadata",
		[]string{"date", "camera", "iso", "aperture"},
		true,
		"Images",
		"A JSON array of EXIF/image metadata fields to display in image detail views.",
	),
}

// SeedDefaultSettings inserts initial default settings into the database if they don't already exist.
func SeedDefaultSettings(db *gorm.DB, logger *slog.Logger) {
	for _, setting := range defaultSettings {
		var existing entities.SettingDefault

		// Manually find first to allow Unscoped to find soft-deleted records reliably
		err := db.Unscoped().Where("name = ?", setting.Name).First(&existing).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Not found, create it
				if createErr := db.Create(&setting).Error; createErr != nil {
					logger.Error("failed to create default setting", slog.String("setting_name", setting.Name), slog.Any("error", createErr))
				} else {
					logger.Info("created default setting", slog.String("setting_name", setting.Name))
				}
			} else {
				logger.Error("failed to query default setting", slog.String("setting_name", setting.Name), slog.Any("error", err))
			}
			continue
		}

		// If found (including soft-deleted), check if update is needed
		wasDeleted := existing.DeletedAt.Valid

		if wasDeleted ||
			existing.Value != setting.Value ||
			existing.Description != setting.Description ||
			existing.Group != setting.Group ||
			existing.IsUserEditable != setting.IsUserEditable ||
			existing.ValueType != setting.ValueType ||
			!utils.EqualStringSlices(existing.AllowedValues, setting.AllowedValues) {

			existing.Value = setting.Value
			existing.Description = setting.Description
			existing.Group = setting.Group
			existing.IsUserEditable = setting.IsUserEditable
			existing.ValueType = setting.ValueType
			existing.AllowedValues = setting.AllowedValues

			if wasDeleted {
				existing.DeletedAt = gorm.DeletedAt{} // Reset to NULL
			}

			if updateErr := db.Unscoped().Save(&existing).Error; updateErr != nil {
				logger.Error("failed to update default setting", slog.String("setting_name", setting.Name), slog.Any("error", updateErr))
			} else {
				logger.Info("updated default setting", slog.String("setting_name", setting.Name))
			}
		} else {
			logger.Info("default setting up-to-date", slog.String("setting_name", setting.Name))
		}
	}
}
