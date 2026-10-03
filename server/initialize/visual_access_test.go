package initialize

import (
	adapter "github.com/casbin/gorm-adapter/v3"
	"github.com/glebarez/sqlite"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	"gorm.io/gorm"
	"testing"
)

func TestVisualPermissionMigration(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(&aiModel.ImageConfig{}, &adapter.CasbinRule{}, &system.SysApi{}); err != nil {
		t.Fatal(err)
	}
	rules := []adapter.CasbinRule{
		{Ptype: "p", V0: "viewer", V1: "/ai/modelConfig/list", V2: "GET"},
		{Ptype: "p", V0: "editor", V1: "/ai/modelConfig", V2: "PUT"},
		{Ptype: "p", V0: "writer", V1: "/blog/ai/chat", V2: "POST"},
		{Ptype: "p", V0: "uploader", V1: "/fileUploadAndDownload/upload", V2: "POST"},
	}
	if err = db.Create(&rules).Error; err != nil {
		t.Fatal(err)
	}
	if err = migrateVisualAccess(db); err != nil {
		t.Fatal(err)
	}
	check := func(role, path, method string, want int64) {
		t.Helper()
		var count int64
		if err := db.Model(&adapter.CasbinRule{}).Where("v0 = ? AND v1 = ? AND v2 = ?", role, path, method).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != want {
			t.Fatalf("%s %s %s = %d, want %d", role, method, path, count, want)
		}
	}
	check("viewer", "/ai/modelConfig/image", "GET", 1)
	check("viewer", "/ai/modelConfig/image", "PUT", 0)
	check("viewer", "/ai/modelConfig/image/test", "POST", 0)
	check("editor", "/ai/modelConfig/image/test", "POST", 1)
	check("writer", "/blog/ai/visual/generate", "POST", 1)
	check("writer", "/blog/ai/visual/adopt", "POST", 0)
	check("uploader", "/blog/ai/visual/adopt", "POST", 1)
	if err = db.Where("v1 = ?", "/blog/ai/visual/generate").Delete(&adapter.CasbinRule{}).Error; err != nil {
		t.Fatal(err)
	}
	if err = migrateVisualAccess(db); err != nil {
		t.Fatal(err)
	}
	check("writer", "/blog/ai/visual/generate", "POST", 0)
	db.Where("v1 = ?", "/ai/modelConfig/image/test").Delete(&adapter.CasbinRule{})
	if err = migrateVisualAccess(db); err != nil {
		t.Fatal(err)
	}
	check("editor", "/ai/modelConfig/image/test", "POST", 0)
}
