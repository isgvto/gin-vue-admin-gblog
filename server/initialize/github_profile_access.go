package initialize

import (
	adapter "github.com/casbin/gorm-adapter/v3"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	"gorm.io/gorm"
)

// The API catalog row is the one-time migration marker. Revoked policies are
// never reintroduced on subsequent startups.
func migrateGitHubProfileAccess(db *gorm.DB) error {
	if !db.Migrator().HasTable(&adapter.CasbinRule{}) || !db.Migrator().HasTable(&system.SysApi{}) {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		mappings := []struct{ from, method, description string }{{"/user/getUserInfo", "GET", "读取自己的 GitHub 展示数据"}, {"/user/setSelfInfo", "PUT", "设置自己的 GitHub 展示账号"}}
		for _, mapping := range mappings {
			var count int64
			if err := tx.Model(&system.SysApi{}).Where("path = ? AND method = ?", "/user/github", mapping.method).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				continue
			}
			var rules []adapter.CasbinRule
			if err := tx.Where("ptype = ? AND v1 = ? AND v2 = ?", "p", mapping.from, mapping.method).Find(&rules).Error; err != nil {
				return err
			}
			for _, rule := range rules {
				rule.ID = 0
				rule.V1 = "/user/github"
				if err := tx.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", rule.Ptype, rule.V0, rule.V1, rule.V2).FirstOrCreate(&rule).Error; err != nil {
					return err
				}
			}
			if err := tx.Create(&system.SysApi{Path: "/user/github", Method: mapping.method, ApiGroup: "系统用户", Description: mapping.description}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
