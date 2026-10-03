package initialize

import (
	adapter "github.com/casbin/gorm-adapter/v3"
	"github.com/glebarez/sqlite"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	"gorm.io/gorm"
	"testing"
)

func TestGitHubProfilePermissionMigration(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(&adapter.CasbinRule{}, &system.SysApi{}); err != nil {
		t.Fatal(err)
	}
	db.Create(&adapter.CasbinRule{Ptype: "p", V0: "reader", V1: "/user/getUserInfo", V2: "GET"})
	db.Create(&adapter.CasbinRule{Ptype: "p", V0: "editor", V1: "/user/setSelfInfo", V2: "PUT"})
	if err = migrateGitHubProfileAccess(db); err != nil {
		t.Fatal(err)
	}
	var rules []adapter.CasbinRule
	db.Where("v1 = ?", "/user/github").Find(&rules)
	if len(rules) != 2 {
		t.Fatalf("expected inherited permissions: %+v", rules)
	}
	for _, rule := range rules {
		if (rule.V0 == "reader" && rule.V2 != "GET") || (rule.V0 == "editor" && rule.V2 != "PUT") {
			t.Fatal("permissions widened")
		}
	}
	db.Where("v1 = ?", "/user/github").Delete(&adapter.CasbinRule{})
	if err = migrateGitHubProfileAccess(db); err != nil {
		t.Fatal(err)
	}
	var count int64
	db.Model(&adapter.CasbinRule{}).Where("v1 = ?", "/user/github").Count(&count)
	if count != 0 {
		t.Fatal("revoked permissions restored")
	}
}
