package initialize

import (
	"testing"

	adapter "github.com/casbin/gorm-adapter/v3"
	"github.com/glebarez/sqlite"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	"gorm.io/gorm"
)

func TestErrorAnalysisAccessMigration(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&adapter.CasbinRule{}, &system.SysApi{}, &aiModel.ErrorAnalysisConfig{}); err != nil {
		t.Fatal(err)
	}
	policies := []adapter.CasbinRule{
		{Ptype: "p", V0: "reader", V1: "/ai/modelConfig/list", V2: "GET"},
		{Ptype: "p", V0: "editor", V1: "/ai/modelConfig", V2: "PUT"},
		{Ptype: "p", V0: "analyst", V1: "/sysError/getSysErrorSolution", V2: "GET"},
	}
	if err := db.Create(&policies).Error; err != nil {
		t.Fatal(err)
	}
	if err := migrateErrorAnalysisAccess(db); err != nil {
		t.Fatal(err)
	}
	var count int64
	check := func(role, path, method string, expected int64) {
		t.Helper()
		if err := db.Model(&adapter.CasbinRule{}).Where("v0 = ? AND v1 = ? AND v2 = ?", role, path, method).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != expected {
			t.Fatalf("%s %s %s: got %d", role, method, path, count)
		}
	}
	check("reader", "/ai/modelConfig/errorAnalysis", "GET", 1)
	check("reader", "/ai/modelConfig/errorAnalysis", "PUT", 0)
	check("editor", "/ai/modelConfig/errorAnalysis", "PUT", 1)
	check("analyst", "/sysError/getSysErrorSolution", "POST", 1)
	check("analyst", "/sysError/getSysErrorSolution", "GET", 0)
	if err := db.Where("v0 = ? AND v1 = ?", "editor", "/ai/modelConfig/errorAnalysis").Delete(&adapter.CasbinRule{}).Error; err != nil {
		t.Fatal(err)
	}
	if err := migrateErrorAnalysisAccess(db); err != nil {
		t.Fatal(err)
	}
	check("editor", "/ai/modelConfig/errorAnalysis", "PUT", 0)
}
