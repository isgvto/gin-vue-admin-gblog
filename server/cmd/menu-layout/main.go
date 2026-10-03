package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/isgvto/gin-vue-admin-gblog/server/core"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	"github.com/isgvto/gin-vue-admin-gblog/server/initialize"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	menuService "github.com/isgvto/gin-vue-admin-gblog/server/service/system"
	"go.uber.org/zap"
)

func main() {
	global.GVA_LOG = zap.NewNop()
	global.GVA_VP = core.Viper()
	db := initialize.Gorm()
	if db == nil {
		fmt.Fprintln(os.Stderr, "database unavailable")
		os.Exit(1)
	}
	var menus []system.SysBaseMenu
	if err := db.Preload("Parameters").Preload("MenuBtn").Order("id").Find(&menus).Error; err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if backupPath := os.Getenv("MENU_LAYOUT_VERIFY"); backupPath != "" {
		var backup struct {
			Menus []system.SysBaseMenu `json:"menus"`
			Links []struct {
				MenuID      string `json:"menuId"`
				AuthorityID string `json:"authorityId"`
			} `json:"authorityMenus"`
		}
		data, err := os.ReadFile(backupPath)
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(data, &backup); err != nil {
			panic(err)
		}
		oldMap, newMap := map[uint]system.SysBaseMenu{}, map[uint]system.SysBaseMenu{}
		for _, m := range backup.Menus {
			oldMap[m.ID] = m
		}
		for _, m := range menus {
			newMap[m.ID] = m
		}
		var fullPath func(map[uint]system.SysBaseMenu, uint) string
		fullPath = func(items map[uint]system.SysBaseMenu, id uint) string {
			m := items[id]
			if strings.HasPrefix(m.Path, "/") || strings.HasPrefix(m.Path, "http") {
				return m.Path
			}
			prefix := "/layout"
			if m.ParentId != 0 {
				prefix = fullPath(items, m.ParentId)
			}
			return path.Join(prefix, m.Path)
		}
		hiddenCount := 0
		for _, old := range backup.Menus {
			now, exists := newMap[old.ID]
			if !exists || now.Hidden != old.Hidden || now.Name != old.Name || now.Component != old.Component || fullPath(oldMap, old.ID) != fullPath(newMap, old.ID) || len(now.Parameters) != len(old.Parameters) || len(now.MenuBtn) != len(old.MenuBtn) {
				panic(fmt.Sprintf("menu contract changed: %s", old.Name))
			}
			if old.Hidden {
				hiddenCount++
				if now.ParentId != old.ParentId || now.Title != old.Title || now.Sort != old.Sort {
					panic("hidden menu changed")
				}
			}
		}
		var links []system.SysAuthorityMenu
		if err := db.Find(&links).Error; err != nil {
			panic(err)
		}
		current := map[string]bool{}
		for _, link := range links {
			current[link.AuthorityId+":"+link.MenuId] = true
		}
		for _, link := range backup.Links {
			if !current[link.AuthorityID+":"+link.MenuID] {
				panic("existing role permission removed")
			}
		}
		fmt.Printf("Verified %d original menus, %d hidden menus, %d existing role links; all original URLs preserved.\n", len(backup.Menus), hiddenCount, len(backup.Links))
		return
	}
	if os.Getenv("MENU_LAYOUT_APPLY") == "1" {
		backupPath := os.Getenv("MENU_LAYOUT_BACKUP")
		if backupPath == "" {
			fmt.Fprintln(os.Stderr, "MENU_LAYOUT_BACKUP is required")
			os.Exit(1)
		}
		var links []system.SysAuthorityMenu
		if err := db.Find(&links).Error; err != nil {
			panic(err)
		}
		backupLinks := make([]map[string]string, 0, len(links))
		for _, link := range links {
			backupLinks = append(backupLinks, map[string]string{"menuId": link.MenuId, "authorityId": link.AuthorityId})
		}
		backup, err := json.MarshalIndent(map[string]any{"createdAt": time.Now(), "menus": menus, "authorityMenus": backupLinks}, "", "  ")
		if err != nil {
			panic(err)
		}
		file, err := os.OpenFile(backupPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			panic(err)
		}
		if _, err = file.Write(backup); err != nil {
			file.Close()
			panic(err)
		}
		if err := file.Close(); err != nil {
			panic(err)
		}
		if err := menuService.OrganizeMenuLayout(db); err != nil {
			panic(err)
		}
		fmt.Println("Menu layout updated. Backup:", backupPath)
		if err := db.Order("parent_id, sort, id").Find(&menus).Error; err != nil {
			panic(err)
		}
	}
	for _, m := range menus {
		data, _ := json.Marshal(map[string]any{"id": m.ID, "parent": m.ParentId, "name": m.Name, "title": m.Title, "path": m.Path, "hidden": m.Hidden, "sort": m.Sort, "component": m.Component})
		fmt.Println(string(data))
	}
}
