package system

import (
	"testing"

	"github.com/glebarez/sqlite"
	model "github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	"gorm.io/gorm"
)

func TestMenuLayoutPreservesRoutesVisibilityAndPermissions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.SysBaseMenu{}, &model.SysAuthorityMenu{}); err != nil {
		t.Fatal(err)
	}
	roots := []model.SysBaseMenu{
		{Name: "gblog", Path: "gblog", Component: "view/routerHolder.vue"},
		{Name: "superAdmin", Path: "admin", Component: "view/superAdmin/index.vue"},
		{Name: "systemTools", Path: "systemTools", Component: "view/systemTools/index.vue"},
		{Name: "ai", Path: "ai", Component: "view/routerHolder.vue"},
	}
	if err := db.Create(&roots).Error; err != nil {
		t.Fatal(err)
	}
	menus := []model.SysBaseMenu{
		{Name: "list", Path: "list", ParentId: roots[0].ID, Component: "view/blog/blog/BlogList.vue"},
		{Name: "edit", Path: "edit/:id?", ParentId: roots[0].ID, Component: "view/blog/blog/WriteBlog.vue"},
		{Name: "menu", Path: "menu", ParentId: roots[1].ID, Component: "view/superAdmin/menu/menu.vue"},
		{Name: "modelConfig", Path: "modelConfig", ParentId: roots[1].ID, Component: "view/ai/modelConfig/modelConfig.vue"},
		{Name: "operationLog", Path: "operationLog", ParentId: roots[0].ID, Component: "view/blogLog/OperationLog.vue"},
		{Name: "autoCodeEdit", Path: "autoCodeEdit/:id", ParentId: roots[2].ID, Hidden: true, Meta: model.Meta{Title: "hidden unchanged"}},
		{Name: "blogAbout", Path: "blogAbout", ParentId: roots[0].ID, Hidden: true, Meta: model.Meta{Title: "hidden about"}},
		{Name: "custom", Path: "custom", ParentId: roots[0].ID},
	}
	if err := db.Create(&menus).Error; err != nil {
		t.Fatal(err)
	}
	link := model.SysAuthorityMenu{MenuId: "5", AuthorityId: "100"}
	if menus[0].ID != 5 {
		t.Fatal("unexpected fixture ID")
	}
	if err := db.Create(&link).Error; err != nil {
		t.Fatal(err)
	}
	legacyFolder := model.SysBaseMenu{Name: "navArticles", Path: "navArticles", ParentId: roots[0].ID, Component: "view/routerHolder.vue"}
	if err := db.Create(&legacyFolder).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&menus[0]).Updates(map[string]any{"parent_id": legacyFolder.ID, "path": "/layout/gblog/list"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := OrganizeMenuLayout(db); err != nil {
		t.Fatal(err)
	}
	var actual []model.SysBaseMenu
	db.Order("id").Find(&actual)
	byName := map[string]model.SysBaseMenu{}
	for _, m := range actual {
		byName[m.Name] = m
	}
	if _, exists := byName["navArticles"]; exists {
		t.Fatal("intermediate folder still visible")
	}
	var retired model.SysBaseMenu
	if err := db.Unscoped().First(&retired, legacyFolder.ID).Error; err != nil || !retired.DeletedAt.Valid {
		t.Fatal("old folder is not recoverable", err)
	}
	for name, expected := range map[string]string{"list": "/layout/gblog/list", "edit": "/layout/gblog/edit/:id?", "menu": "/layout/admin/menu", "modelConfig": "/layout/admin/modelConfig", "operationLog": "/layout/gblog/operationLog"} {
		if byName[name].Path != expected {
			t.Errorf("%s path = %s", name, byName[name].Path)
		}
	}
	for _, index := range []int{5, 6, 7} {
		before, after := menus[index], byName[menus[index].Name]
		if before.ParentId != after.ParentId || before.Path != after.Path || before.Hidden != after.Hidden || before.Title != after.Title {
			t.Errorf("unexpected change to %s", before.Name)
		}
	}
	var links []model.SysAuthorityMenu
	db.Find(&links)
	for _, link := range links {
		if link.MenuId != "5" && link.MenuId != "1" {
			t.Errorf("granted unrelated menu: %+v", link)
		}
	}
	if byName["list"].ParentId != roots[0].ID || byName["modelConfig"].ParentId != roots[3].ID {
		t.Fatal("features are not direct second-level menus")
	}
	count := len(actual)
	linkCount := len(links)
	if err := OrganizeMenuLayout(db); err != nil {
		t.Fatal(err)
	}
	db.Find(&actual)
	db.Find(&links)
	if len(actual) != count || len(links) != linkCount {
		t.Fatal("rerun duplicated folders or roles")
	}
}
