package entities

import (
	"fmt"

	"gorm.io/gorm"

	"imagine/internal/dto"
)

func CountUsers(db *gorm.DB) (int64, error) {
	var count int64
	if err := db.Model(&User{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

func CountSuperadmins(db *gorm.DB) (int64, error) {
	var count int64
	if err := db.Model(&User{}).Where("role = ?", dto.UserRoleSuperadmin).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count superadmins: %w", err)
	}

	return count, nil
}

// HardDeleteUser permanently removes a user and all associated data (sessions, settings, onboarding).
// This operation is irreversible.
func HardDeleteUser(db *gorm.DB, userUid string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 1. Delete all sessions for the user
		if err := tx.Unscoped().Where("user_uid = ?", userUid).Delete(&Session{}).Error; err != nil {
			return fmt.Errorf("failed to delete user sessions: %w", err)
		}

		// 2. Delete all setting overrides (includes onboarding status)
		// Note: SettingOverride uses UserId field for the user's UID
		if err := tx.Unscoped().Where("user_id = ?", userUid).Delete(&SettingOverride{}).Error; err != nil {
			return fmt.Errorf("failed to delete user settings: %w", err)
		}

		// 3. Delete the user record itself
		if err := tx.Unscoped().Where("uid = ?", userUid).Delete(&User{}).Error; err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		return nil
	})
}