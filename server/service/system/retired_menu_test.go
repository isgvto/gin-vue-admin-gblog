package system

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common"
	"gorm.io/gorm"
)

func TestRemovedAIPageDrawingMenu(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	if err := db.Exec("CREATE TABLE menus (id INTEGER PRIMARY KEY, name TEXT, component TEXT)").Error; err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct{ name, component string }{
		{"renamedDrawing", "view/systemTools/autoCode/picture.vue"},
		{"picture", "view/blog/article/index.vue"},
		{"autoCode", "view/systemTools/autoCode/index.vue"},
	} {
		if err := db.Exec("INSERT INTO menus (name, component) VALUES (?, ?)", row.name, row.component).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Exec("INSERT INTO menus (name, component) VALUES ('parent', NULL)").Error; err != nil {
		t.Fatal(err)
	}
	var names []string
	if err := db.Table("menus").Scopes(excludeAIPageDrawingMenu).Order("id").Pluck("name", &names).Error; err != nil {
		t.Fatal(err)
	}
	if len(names) != 3 || names[0] != "picture" || names[1] != "autoCode" || names[2] != "parent" {
		t.Fatalf("unexpected remaining menus: %v", names)
	}
	var count int64
	if err := db.Table("menus").Scopes(excludeAIPageDrawingMenu).Where("name = ?", "renamedDrawing").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("removed menu still eligible as default route: count=%d err=%v", count, err)
	}
}

func TestRemovedAIPageDrawingLLMMode(t *testing.T) {
	original := global.GVA_CONFIG.AutoCode.AiPath
	t.Cleanup(func() { global.GVA_CONFIG.AutoCode.AiPath = original })
	global.GVA_CONFIG.AutoCode.AiPath = "https://example.invalid/{FUNC}"
	for _, mode := range []string{"createWeb", " createWeb "} {
		if _, err := buildLLMAutoPath(common.JSONMap{"mode": mode}); err == nil {
			t.Fatalf("removed mode %q was accepted", mode)
		}
	}
	path, err := buildLLMAutoPath(common.JSONMap{"mode": "apiCompletion"})
	if err != nil || path != "https://example.invalid/apiCompletion" {
		t.Fatalf("shared AI mode affected: path=%q err=%v", path, err)
	}
}

func TestWorkflowModesNoLongerUseAIPath(t *testing.T) {
	previous := global.GVA_CONFIG.AutoCode.AiPath
	global.GVA_CONFIG.AutoCode.AiPath = "https://example.invalid/{FUNC}"
	t.Cleanup(func() { global.GVA_CONFIG.AutoCode.AiPath = previous })
	for _, mode := range []string{"analysisChat", "workflowPromptChat", " analysisChat "} {
		if _, err := buildLLMAutoPath(common.JSONMap{"mode": mode}); err == nil {
			t.Fatalf("legacy workflow mode accepted: %s", mode)
		}
	}
}
