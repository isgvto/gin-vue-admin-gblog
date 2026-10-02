package initialize

import (
	adapter "github.com/casbin/gorm-adapter/v3"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	"gorm.io/gorm"
)

func migrateWorkflowAccess(db *gorm.DB) error {
	if !db.Migrator().HasTable(&adapter.CasbinRule{}) {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var cfg aiModel.WorkflowConfig
		if err := tx.Where("id = ?", 1).Attrs(aiModel.WorkflowConfig{ID: 1, TimeoutSeconds: 180}).FirstOrCreate(&cfg).Error; err != nil {
			return err
		}
		if cfg.PermissionsMigrated {
			return nil
		}
		mappings := []struct{ from, method, to, targetMethod, description string }{
			{"/ai/modelConfig/list", "GET", "/ai/modelConfig/workflow", "GET", "读取需求工作流模型分配"},
			{"/ai/modelConfig", "PUT", "/ai/modelConfig/workflow", "PUT", "保存需求工作流模型分配"},
			{"/ai/modelConfig/testConnection", "POST", "/ai/modelConfig/workflow/test", "POST", "测试需求工作流模型"},
			{"/autoCode/saveAIWorkflowSession", "POST", "/autoCode/aiWorkflowChat", "POST", "AI 需求工作流对话"},
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
			api := system.SysApi{Path: m.to, Method: m.targetMethod, ApiGroup: "AI 功能", Description: m.description}
			if err := tx.Where("path = ? AND method = ?", api.Path, api.Method).FirstOrCreate(&api).Error; err != nil {
				return err
			}
		}
		return tx.Model(&cfg).Update("permissions_migrated", true).Error
	})
}
