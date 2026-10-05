package blog

import (
	"encoding/json"

	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	"gorm.io/gorm"
)

// Only these retired keys are removed; site text, social links and badge content stay intact.
var retiredSiteSettings = []string{"bg1", "bg2", "bg3", "playlistServer", "playlistId"}

func isRetiredSiteSetting(name *string) bool {
	if name == nil {
		return false
	}
	for _, key := range retiredSiteSettings {
		if *name == key {
			return true
		}
	}
	return false
}

func stripBadgeColor(value *string) (*string, bool) {
	if value == nil {
		return value, false
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(*value), &fields); err != nil {
		// Keep malformed legacy values available for correction in the settings form.
		return value, false
	}
	if _, exists := fields["color"]; !exists {
		return value, false
	}
	delete(fields, "color")
	encoded, err := json.Marshal(fields)
	if err != nil {
		return value, false
	}
	cleaned := string(encoded)
	return &cleaned, true
}

func activeSiteSettings(rows []blogModel.SiteSetting) []blogModel.SiteSetting {
	result := make([]blogModel.SiteSetting, 0, len(rows))
	for _, row := range rows {
		if isRetiredSiteSetting(row.NameEn) {
			continue
		}
		if row.Type != nil && *row.Type == 3 {
			row.Value, _ = stripBadgeColor(row.Value)
		}
		result = append(result, row)
	}
	return result
}

// CleanupRetiredSiteSettings is idempotent and also handles installations upgrading from old data.
func CleanupRetiredSiteSettings(db *gorm.DB) error {
	return db.Transaction(cleanupRetiredSiteSettings)
}

func cleanupRetiredSiteSettings(tx *gorm.DB) error {
	if err := tx.Where("name_en IN ?", retiredSiteSettings).Delete(&blogModel.SiteSetting{}).Error; err != nil {
		return err
	}
	var badges []blogModel.SiteSetting
	if err := tx.Where("type = ?", 3).Find(&badges).Error; err != nil {
		return err
	}
	for _, badge := range badges {
		if value, changed := stripBadgeColor(badge.Value); changed {
			if err := tx.Model(&blogModel.SiteSetting{}).Where("id = ?", badge.ID).Update("value", value).Error; err != nil {
				return err
			}
		}
	}
	if tx.Migrator().HasTable(&blogModel.About{}) {
		return tx.Where("name_en = ?", "musicId").Delete(&blogModel.About{}).Error
	}
	return nil
}
