package initialize

import (
	"errors"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	"gorm.io/gorm"
)

// Copy the former shared assignment once, preserving shared model records.
func migrateImageModelBinding(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var cfg aiModel.ImageConfig
		if err := tx.First(&cfg, 1).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		} else if err != nil {
			return err
		}
		if cfg.IndependentMigrated {
			return nil
		}
		updates := map[string]any{"independent_migrated": true, "model_id": 0}
		if cfg.Model == "" && cfg.APIKey == "" && (cfg.ModelID != 0 || cfg.Enabled) {
			var model aiModel.AiModelConfig
			query := tx.Where("id = ?", cfg.ModelID)
			if cfg.ModelID == 0 {
				query = tx.Where("is_default = ? AND status = ?", true, true)
			}
			err := query.First(&model).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				updates["enabled"] = false
			} else if err != nil {
				return err
			} else {
				updates["provider"] = model.Provider
				updates["base_url"] = model.BaseURL
				updates["model"] = model.Model
				updates["api_key"] = model.APIKey
				if model.Provider != "ark" && model.Provider != "openai" {
					updates["provider"] = "openai"
				}
				if !model.Status {
					updates["enabled"] = false
				}
			}
		}
		return tx.Model(&cfg).Updates(updates).Error
	})
}
