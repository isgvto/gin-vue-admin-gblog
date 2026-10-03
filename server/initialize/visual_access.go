package initialize

import (
	adapter "github.com/casbin/gorm-adapter/v3"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	"gorm.io/gorm"
)

func migrateVisualAccess(db *gorm.DB) error {
	if !db.Migrator().HasTable(&adapter.CasbinRule{}) {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		cfg := aiModel.ImageConfig{ID: 1, TimeoutSeconds: 180}
		if err := tx.FirstOrCreate(&cfg, "id = ?", 1).Error; err != nil {
			return err
		}
		if !cfg.TestAccessMigrated {
			var rules []adapter.CasbinRule
			if err := tx.Where("ptype = ? AND v1 = ? AND v2 = ?", "p", "/ai/modelConfig/image", "PUT").Find(&rules).Error; err != nil {
				return err
			}
			// First installation may not yet have image PUT; derive from model editing.
			if !cfg.PermissionsMigrated {
				if err := tx.Where("ptype = ? AND v1 = ? AND v2 = ?", "p", "/ai/modelConfig", "PUT").Find(&rules).Error; err != nil {
					return err
				}
			}
			for _, rule := range rules {
				rule.ID = 0
				rule.V1 = "/ai/modelConfig/image/test"
				rule.V2 = "POST"
				if err := tx.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", rule.Ptype, rule.V0, rule.V1, rule.V2).FirstOrCreate(&rule).Error; err != nil {
					return err
				}
			}
			api := system.SysApi{Path: "/ai/modelConfig/image/test", Method: "POST", ApiGroup: "AI 配图", Description: "测试独立图片模型"}
			if err := tx.Where("path = ? AND method = ?", api.Path, api.Method).FirstOrCreate(&api).Error; err != nil {
				return err
			}
			if err := tx.Model(&cfg).Update("test_access_migrated", true).Error; err != nil {
				return err
			}
		}
		if cfg.PermissionsMigrated {
			return nil
		}
		mappings := []struct{ from, method, to, targetMethod, description string }{
			{"/ai/modelConfig/list", "GET", "/ai/modelConfig/image", "GET", "读取图片生成模型配置"},
			{"/ai/modelConfig", "PUT", "/ai/modelConfig/image", "PUT", "保存图片生成模型配置"},
			{"/blog/ai/status", "GET", "/blog/ai/visual/status", "GET", "读取配图状态"},
			{"/blog/ai/chat", "POST", "/blog/ai/visual/plan", "POST", "推荐配图类型与描述"},
			{"/blog/ai/chat", "POST", "/blog/ai/visual/generate", "POST", "生成配图预览"},
			{"/blog/ai/chat", "POST", "/blog/ai/visual/discard", "POST", "丢弃配图预览"},
			{"/fileUploadAndDownload/upload", "POST", "/blog/ai/visual/adopt", "POST", "采用配图并上传"},
		}
		for _, m := range mappings {
			var rules []adapter.CasbinRule
			if err := tx.Where("ptype = ? AND v1 = ? AND v2 = ?", "p", m.from, m.method).Find(&rules).Error; err != nil {
				return err
			}
			for _, rule := range rules {
				rule.ID = 0
				rule.V1 = m.to
				rule.V2 = m.targetMethod
				if err := tx.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", rule.Ptype, rule.V0, rule.V1, rule.V2).FirstOrCreate(&rule).Error; err != nil {
					return err
				}
			}
			api := system.SysApi{Path: m.to, Method: m.targetMethod, ApiGroup: "AI 配图", Description: m.description}
			if err := tx.Where("path = ? AND method = ?", api.Path, api.Method).FirstOrCreate(&api).Error; err != nil {
				return err
			}
		}
		return tx.Model(&cfg).Update("permissions_migrated", true).Error
	})
}
