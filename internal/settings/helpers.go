package settings

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"imagine/internal/dto"
	"imagine/internal/entities"

	"gorm.io/gorm"
)

// CleanupSettingDefaults removes duplicate entries from the setting_defaults table
// to prepare for unique constraints. It keeps the entry with the smallest ID.
func CleanupSettingDefaults(db *gorm.DB, logger *slog.Logger) {
	var duplicates []struct {
		Name  string
		Count int
	}

	// Find names that have more than one entry (including soft-deleted ones)
	db.Unscoped().Model(&entities.SettingDefault{}).
		Select("name, count(*)").
		Group("name").
		Having("count(*) > ?", 1).
		Find(&duplicates)

	if len(duplicates) == 0 {
		logger.Info("no duplicate setting_defaults found, skipping cleanup")
	} else {
		for _, dup := range duplicates {
			logger.Warn("found duplicate setting_default entries, cleaning up...", slog.String("setting_name", dup.Name), slog.Int("count", dup.Count))

			var existingSettings []entities.SettingDefault
			db.Unscoped().Where("name = ?", dup.Name).Order("id ASC").Find(&existingSettings)

			// Keep the first one, delete the rest
			if len(existingSettings) > 1 {
				idsToDelete := make([]uint, 0, len(existingSettings)-1)
				for i := 1; i < len(existingSettings); i++ {
					idsToDelete = append(idsToDelete, existingSettings[i].ID)
				}
				if err := db.Unscoped().Delete(&entities.SettingDefault{}, idsToDelete).Error; err != nil {
					logger.Error("failed to delete duplicate setting_defaults", slog.String("setting_name", dup.Name), slog.Any("error", err))
				} else {
					logger.Info("deleted duplicate setting_defaults", slog.String("setting_name", dup.Name), slog.Int("deleted_count", len(idsToDelete)))
				}
			}
		}
	}
}

// CleanupSettingOverrides removes duplicate entries from the setting_overrides table.
func CleanupSettingOverrides(db *gorm.DB, logger *slog.Logger) {
	var duplicates []struct {
		UserId string
		Name   string
		Count  int
	}

	// Find duplicates based on user_id and name
	db.Unscoped().Model(&entities.SettingOverride{}).
		Select("user_id, name, count(*)").
		Group("user_id, name").
		Having("count(*) > ?", 1).
		Find(&duplicates)

	if len(duplicates) == 0 {
		logger.Info("no duplicate setting_overrides found, skipping cleanup")
		return
	}

	for _, dup := range duplicates {
		logger.Warn("found duplicate setting_override entries, cleaning up...", slog.String("user_id", dup.UserId), slog.String("name", dup.Name), slog.Int("count", dup.Count))

		var existingOverrides []entities.SettingOverride
		db.Unscoped().Where("user_id = ? AND name = ?", dup.UserId, dup.Name).Order("id ASC").Find(&existingOverrides)

		if len(existingOverrides) > 1 {
			idsToDelete := make([]uint, 0, len(existingOverrides)-1)
			for i := 1; i < len(existingOverrides); i++ {
				idsToDelete = append(idsToDelete, existingOverrides[i].ID)
			}
			if err := db.Unscoped().Delete(&entities.SettingOverride{}, idsToDelete).Error; err != nil {
				logger.Error("failed to delete duplicate setting_overrides", slog.String("user_id", dup.UserId), slog.String("name", dup.Name), slog.Any("error", err))
			} else {
				logger.Info("deleted duplicate setting_overrides", slog.String("user_id", dup.UserId), slog.String("name", dup.Name), slog.Int("deleted_count", len(idsToDelete)))
			}
		}
	}
}

// SetSetting creates or updates a setting in the database.
// If userID is provided, it attempts to set a SettingOverride for that user.
// Otherwise, it sets a SettingDefault.
func SetSetting(db *gorm.DB, name string, value string, userID *string) error {
	if userID != nil && *userID != "" {
		// Try to create/update SettingOverride
		var override entities.SettingOverride
		result := db.Where("name = ? AND user_id = ?", name, *userID).First(&override)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// Create new override
				override = entities.SettingOverride{
					Name:   name,
					UserId: *userID,
					Value:  value,
				}
				if createErr := db.Create(&override).Error; createErr != nil {
					return fmt.Errorf("failed to create user setting override: %w", createErr)
				}
				return nil
			}
			return fmt.Errorf("failed to query user setting override: %w", result.Error)
		}

		// Update existing override
		override.Value = value
		if updateErr := db.Save(&override).Error; updateErr != nil {
			return fmt.Errorf("failed to update user setting override: %w", updateErr)
		}
		return nil
	} else {
		// Try to create/update SettingDefault
		var def entities.SettingDefault
		result := db.Where("name = ?", name).First(&def)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// Create new default setting - this scenario should ideally not happen
				// for settings defined in seed.go, but provides robustness.
				def = entities.SettingDefault{
					Name:  name,
					Value: value,
					// Other fields would need to be populated, but for 'first_run_complete'
					// we only care about name and value.
					ValueType:      dto.Boolean, // Assuming for internal flags
					IsUserEditable: false,
					Group:          "System",
					Description:    fmt.Sprintf("Internal setting for %s", name),
				}
				if createErr := db.Create(&def).Error; createErr != nil {
					return fmt.Errorf("failed to create default setting: %w", createErr)
				}
				return nil
			}
			return fmt.Errorf("failed to query default setting: %w", result.Error)
		}

		// Update existing default setting
		def.Value = value
		if updateErr := db.Save(&def).Error; updateErr != nil {
			return fmt.Errorf("failed to update default setting: %w", updateErr)
		}
		return nil
	}
}

// GetSetting retrieves a setting's value, prioritizing a user-specific override if userID is provided.
func GetSetting(db *gorm.DB, name string, userID *string) (string, error) {
	// Try to get user-specific override first if userID is provided
	if userID != nil && *userID != "" {
		var override entities.SettingOverride
		err := db.Where("name = ? AND user_id = ?", name, *userID).First(&override).Error
		if err == nil {
			return override.Value, nil
		}
		if err != gorm.ErrRecordNotFound {
			return "", fmt.Errorf("failed to get user setting override: %w", err)
		}
	}

	// Fallback to default setting
	var def entities.SettingDefault
	err := db.Where("name = ?", name).First(&def).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("setting '%s' not found", name)
		}
		return "", fmt.Errorf("failed to get default setting: %w", err)
	}

	return def.Value, nil
}

// Helper function for boolean settings
func BoolSetting(name string, displayName *string, value bool, isUserEditable bool, group, description string) entities.SettingDefault {
	return entities.SettingDefault{
		Name:           name,
		DisplayName:    displayName,
		Value:          strconv.FormatBool(value),
		ValueType:      dto.Boolean,
		IsUserEditable: isUserEditable,
		Group:          group,
		Description:    description,
	}
}

// Helper function for integer settings
func IntSetting(name string, displayName *string, value int, allowedValues []int, isUserEditable bool, group, description string) entities.SettingDefault {
	var allowedValuesStr *[]string
	if len(allowedValues) > 0 {
		strValues := make([]string, len(allowedValues))
		for i, v := range allowedValues {
			strValues[i] = strconv.Itoa(v)
		}
		allowedValuesStr = &strValues
	}

	return entities.SettingDefault{
		Name:           name,
		DisplayName: displayName,
		Value:          strconv.Itoa(value),
		ValueType:      dto.Integer,
		AllowedValues:  allowedValuesStr,
		IsUserEditable: isUserEditable,
		Group:          group,
		Description:    description,
	}
}

// Helper function for string settings
func StringSetting(name string, displayName *string, value string, isUserEditable bool, group, description string) entities.SettingDefault {
	return entities.SettingDefault{
		Name:           name,
		DisplayName: displayName,
		Value:          value,
		ValueType:      dto.String,
		IsUserEditable: isUserEditable,
		Group:          group,
		Description:    description,
	}
}

// Helper function for enum settings
func EnumSetting(name string, displayName *string, value string, allowedValues []string, isUserEditable bool, group, description string) entities.SettingDefault {
	return entities.SettingDefault{
		Name:           name,
		DisplayName: displayName,
		Value:          value,
		ValueType:      dto.Enum,
		AllowedValues:  &allowedValues,
		IsUserEditable: isUserEditable,
		Group:          group,
		Description:    description,
	}
}

// Helper function for JSON settings
func JsonSetting(name string, displayName *string, value interface{}, isUserEditable bool, group, description string) entities.SettingDefault {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON for setting %s: %v", name, err))
	}

	return entities.SettingDefault{
		Name:           name,
		DisplayName: displayName,
		Value:          string(jsonBytes),
		ValueType:      dto.Json,
		IsUserEditable: isUserEditable,
		Group:          group,
		Description:    description,
	}
}
