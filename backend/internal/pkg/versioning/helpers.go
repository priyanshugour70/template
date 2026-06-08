package versioning

import (
	"time"

	"gorm.io/gorm"
)

func NextVersionNumber(tx *gorm.DB, versionTable string, entityID uint64) (uint, error) {
	var row struct {
		Max *uint `gorm:"column:max_v"`
	}
	err := tx.Table(versionTable).
		Where("entity_id = ?", entityID).
		Select("MAX(version_number) AS max_v").
		Scan(&row).Error
	if err != nil {
		return 0, err
	}
	if row.Max == nil || *row.Max == 0 {
		return 1, nil
	}
	return *row.Max + 1, nil
}

func DeactivatePreviousActive(tx *gorm.DB, versionTable string, entityID uint64) error {
	return tx.Table(versionTable).
		Where("entity_id = ? AND is_active = 1", entityID).
		Updates(map[string]interface{}{"is_active": false}).Error
}

// StatusAndActiveForNewVersion returns (status, is_active) for a new version row.
// When autoApprove is true, previously-active rows for this entity are cleared first.
func StatusAndActiveForNewVersion(tx *gorm.DB, versionTable string, entityID uint64, autoApprove bool) (Status, bool, error) {
	if autoApprove {
		if err := DeactivatePreviousActive(tx, versionTable, entityID); err != nil {
			return StatusDraft, false, err
		}
		return StatusApproved, true, nil
	}
	return StatusDraft, false, nil
}

func ActivateVersion(tx *gorm.DB, versionTable string, versionID uint64) error {
	now := time.Now().UTC()
	return tx.Table(versionTable).
		Where("id = ?", versionID).
		Updates(map[string]interface{}{
			"status":      string(StatusApproved),
			"is_active":   true,
			"approved_at": now,
		}).Error
}

func UpdateBasePointer(tx *gorm.DB, baseTable string, entityID, versionID uint64) error {
	return tx.Table(baseTable).
		Where("id = ?", entityID).
		Update("current_version_id", versionID).Error
}

func SetVersionStatus(tx *gorm.DB, versionTable string, versionID uint64, status Status) error {
	return tx.Table(versionTable).
		Where("id = ?", versionID).
		Update("status", string(status)).Error
}

func RejectVersion(tx *gorm.DB, versionTable string, versionID uint64, rejectedBy *uint64, reason string) error {
	now := time.Now().UTC()
	updates := map[string]interface{}{
		"status":      string(StatusRejected),
		"rejected_at": now,
	}
	if rejectedBy != nil {
		updates["rejected_by"] = *rejectedBy
	}
	if reason != "" {
		updates["reject_reason"] = reason
	}
	return tx.Table(versionTable).Where("id = ?", versionID).Updates(updates).Error
}

func SoftDelete(tx *gorm.DB, baseTable string, entityID uint64) error {
	return tx.Table(baseTable).
		Where("id = ?", entityID).
		Update("is_deleted", true).Error
}

func ApproveAndActivate(tx *gorm.DB, baseTable, versionTable string, entityID, versionID uint64, approvedBy *uint64) error {
	if err := DeactivatePreviousActive(tx, versionTable, entityID); err != nil {
		return err
	}
	now := time.Now().UTC()
	updates := map[string]interface{}{
		"status":      string(StatusApproved),
		"is_active":   true,
		"approved_at": now,
	}
	if approvedBy != nil {
		updates["approved_by"] = *approvedBy
	}
	if err := tx.Table(versionTable).Where("id = ?", versionID).Updates(updates).Error; err != nil {
		return err
	}
	return UpdateBasePointer(tx, baseTable, entityID, versionID)
}
