package initialize

import (
	adapter "github.com/casbin/gorm-adapter/v3"
	"github.com/glebarez/sqlite"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	"gorm.io/gorm"
	"testing"
)

func TestWorkflowPermissionMigration(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&aiModel.WorkflowConfig{}, &adapter.CasbinRule{}, &system.SysApi{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&aiModel.WorkflowConfig{ID: 2, TimeoutSeconds: 60}).Error; err != nil {
		t.Fatal(err)
	}
	rules := []adapter.CasbinRule{{Ptype: "p", V0: "viewer", V1: "/ai/modelConfig/list", V2: "GET"}, {Ptype: "p", V0: "author", V1: "/autoCode/saveAIWorkflowSession", V2: "POST"}}
	if err := db.Create(&rules).Error; err != nil {
		t.Fatal(err)
	}
	if err := migrateWorkflowAccess(db); err != nil {
		t.Fatal(err)
	}
	var cfg aiModel.WorkflowConfig
	if err := db.First(&cfg, 1).Error; err != nil || !cfg.PermissionsMigrated {
		t.Fatalf("singleton migration failed: %v", err)
	}
	var count int64
	check := func(role, path, method string, want int64) {
		t.Helper()
		if err := db.Model(&adapter.CasbinRule{}).Where("v0 = ? AND v1 = ? AND v2 = ?", role, path, method).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != want {
			t.Fatalf("%s %s %s: %d", role, method, path, count)
		}
	}
	check("viewer", "/ai/modelConfig/workflow", "GET", 1)
	check("viewer", "/ai/modelConfig/workflow", "PUT", 0)
	check("viewer", "/autoCode/aiWorkflowChat", "POST", 0)
	check("author", "/autoCode/aiWorkflowChat", "POST", 1)
	if err := db.Where("v1 = ?", "/autoCode/aiWorkflowChat").Delete(&adapter.CasbinRule{}).Error; err != nil {
		t.Fatal(err)
	}
	if err := migrateWorkflowAccess(db); err != nil {
		t.Fatal(err)
	}
	check("author", "/autoCode/aiWorkflowChat", "POST", 0)
}
