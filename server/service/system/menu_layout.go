package system

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	model "github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	"gorm.io/gorm"
)

type menuSection struct {
	Name, Parent, Title, Icon string
	Sort                      int
	Items                     []string
}

var menuSections = []menuSection{
	{"navArticles", "gblog", "文章管理", "document", 1, []string{"list", "edit", "blogCategoryList", "blogTagList"}},
	{"navMoments", "gblog", "动态管理", "chat-dot-square", 2, []string{"moment/list", "moment/edit"}},
	{"navCommunity", "gblog", "评论与友链", "chat-round", 3, []string{"commentList", "blogFriendList"}},
	{"navSite", "gblog", "站点与页面", "setting", 4, []string{"blogSiteSetting", "blogAbout"}},
	{"navAIConfig", "ai", "模型与工作流", "magic-stick", 1, []string{"modelConfig", "aiWorkflow", "skills"}},
	{"navAIDev", "ai", "AI 开发与 MCP", "connection", 2, []string{"Cli", "McpApi", "mcpTest", "mcpTool"}},
	{"navAccounts", "superAdmin", "账号与权限", "user", 1, []string{"user", "authority", "apiToken"}},
	{"navNavigation", "superAdmin", "导航与接口", "tickets", 2, []string{"menu", "api"}},
	{"navSettings", "superAdmin", "配置与字典", "setting", 3, []string{"system", "sysParams", "dictionary"}},
	{"navTraffic", "navOperations", "访问分析", "trend-charts", 1, []string{"blogVisitorStats", "blogVisitLog"}},
	{"navAudit", "navOperations", "操作与登录审计", "monitor", 2, []string{"operationLog", "operation", "loginLog"}},
	{"navRuntime", "navOperations", "运行与维护", "cloudy", 3, []string{"state", "blogExceptionLog", "sysError", "sysVersion"}},
	{"navExamples", "example", "业务示例", "collection", 3, []string{"customer", "anInfo"}},
}

// OrganizeMenuLayout changes navigation only. Existing hidden records, route names,
// components, parameters, buttons and leaf authorization links are preserved.
// Absolute paths keep old bookmarks and hardcoded links working after reparenting.
func OrganizeMenuLayout(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var menus []model.SysBaseMenu
		if err := tx.Find(&menus).Error; err != nil {
			return err
		}
		byName := map[string]*model.SysBaseMenu{}
		byID := map[uint]*model.SysBaseMenu{}
		for i := range menus {
			m := &menus[i]
			byName[m.Name] = m
			byID[m.ID] = m
		}
		originalPaths := map[uint]string{}
		var fullPath func(uint, map[uint]bool) (string, error)
		fullPath = func(id uint, visiting map[uint]bool) (string, error) {
			m := byID[id]
			if m == nil {
				return "", fmt.Errorf("菜单父节点不存在: %d", id)
			}
			if visiting[id] {
				return "", fmt.Errorf("菜单存在循环: %s", m.Name)
			}
			if strings.HasPrefix(m.Path, "/") || strings.HasPrefix(m.Path, "http") {
				return m.Path, nil
			}
			visiting[id] = true
			prefix := "/layout"
			if m.ParentId != 0 {
				var err error
				prefix, err = fullPath(m.ParentId, visiting)
				if err != nil {
					return "", err
				}
			}
			delete(visiting, id)
			return path.Join(prefix, m.Path), nil
		}
		for _, m := range menus {
			p, err := fullPath(m.ID, map[uint]bool{})
			if err != nil {
				return err
			}
			originalPaths[m.ID] = p
		}
		// Only create folders when there are visible, existing features to put in them.
		ensureFolder := func(name, title, icon string, parent uint, sort int) (*model.SysBaseMenu, error) {
			if existing := byName[name]; existing != nil {
				if existing.Hidden {
					return nil, fmt.Errorf("分类菜单 %s 已隐藏，停止整理以保留设置", name)
				}
				return existing, nil
			}
			folder := &model.SysBaseMenu{Name: name, Path: name, ParentId: parent, Sort: sort, Component: "view/routerHolder.vue", Meta: model.Meta{Title: title, Icon: icon}}
			if err := tx.Create(folder).Error; err != nil {
				return nil, err
			}
			byName[name] = folder
			byID[folder.ID] = folder
			return folder, nil
		}
		hasOps := false
		for _, section := range menuSections {
			if section.Parent == "navOperations" {
				for _, name := range section.Items {
					if m := byName[name]; m != nil && !m.Hidden {
						hasOps = true
					}
				}
			}
		}
		if hasOps {
			if _, err := ensureFolder("navOperations", "运维监控", "monitor", 0, 6); err != nil {
				return err
			}
		}
		for _, section := range menuSections {
			parent := byName[section.Parent]
			if parent == nil || parent.Hidden {
				continue
			}
			visible := false
			for _, name := range section.Items {
				if m := byName[name]; m != nil && !m.Hidden {
					visible = true
				}
			}
			if !visible {
				continue
			}
			for index, name := range section.Items {
				m := byName[name]
				if m == nil || m.Hidden {
					continue
				}
				if err := tx.Model(m).Updates(map[string]any{"parent_id": parent.ID, "path": originalPaths[m.ID], "sort": section.Sort*10 + index + 1}).Error; err != nil {
					return err
				}
				m.ParentId = parent.ID
			}
		}
		// Retire the previous intermediate folders only after their children move.
		// Soft deletion keeps them recoverable; hidden or custom populated folders stay intact.
		for _, section := range menuSections {
			folder := byName[section.Name]
			if folder == nil || folder.Hidden || folder.Component != "view/routerHolder.vue" {
				continue
			}
			var children int64
			if err := tx.Model(&model.SysBaseMenu{}).Where("parent_id = ?", folder.ID).Count(&children).Error; err != nil {
				return err
			}
			if children == 0 {
				if err := tx.Delete(folder).Error; err != nil {
					return err
				}
				delete(byID, folder.ID)
				delete(byName, folder.Name)
			}
		}

		roots := []struct {
			Name, Title, Icon string
			Sort              int
		}{
			{"gdashboard", "工作台", "pie-chart", 1}, {"gblog", "内容管理", "document", 2}, {"ai", "AI 与集成", "magic-stick", 3},
			{"example", "资源与示例", "folder", 4}, {"superAdmin", "系统管理", "setting", 5}, {"navOperations", "运维监控", "monitor", 6},
			{"systemTools", "开发工具", "tools", 7}, {"plugin", "插件扩展", "box", 8},
		}
		for _, root := range roots {
			if m := byName[root.Name]; m != nil && !m.Hidden {
				updates := map[string]any{"title": root.Title, "icon": root.Icon, "sort": root.Sort}
				if root.Name == "gdashboard" {
					updates["parent_id"] = 0
					updates["path"] = originalPaths[m.ID]
					m.ParentId = 0
				}
				if err := tx.Model(m).Updates(updates).Error; err != nil {
					return err
				}
			}
		}
		titles := map[string]string{"list": "文章列表", "edit": "新建 / 编辑文章", "blogCategoryList": "文章分类", "moment/list": "动态列表", "moment/edit": "新建 / 编辑动态", "blogFriendList": "友情链接", "blogAbout": "关于页面", "upload": "媒体库", "breakpoint": "大文件上传", "customer": "客户管理示例", "anInfo": "公告管理示例", "api": "API 管理", "apiToken": "API 访问令牌", "operation": "系统操作记录", "operationLog": "博客操作日志", "sysError": "系统错误日志", "blogExceptionLog": "博客异常日志", "mcpTest": "MCP 服务管理", "mcpTool": "MCP 工具模板", "McpApi": "AI MCP 构建", "Cli": "AI CLI 构建", "skills": "Skills 管理", "autoPkg": "代码模板配置", "autoCodeAdmin": "生成代码管理"}
		for name, title := range titles {
			if m := byName[name]; m != nil && !m.Hidden {
				if err := tx.Model(m).Update("title", title).Error; err != nil {
					return err
				}
			}
		}
		for _, items := range [][]string{{"upload", "breakpoint"}, {"autoCode", "autoCodeAdmin", "autoPkg", "formCreate", "exportTemplate"}, {"https://plugin.gin-vue-admin.com/", "installPlugin", "pubPlug", "plugin-email"}} {
			for index, name := range items {
				if m := byName[name]; m != nil && !m.Hidden {
					if err := tx.Model(m).Update("sort", index+1).Error; err != nil {
						return err
					}
				}
			}
		}
		// Grant only ancestor folders of already-authorized menus, never another leaf.
		var links []model.SysAuthorityMenu
		if err := tx.Find(&links).Error; err != nil {
			return err
		}
		existing := map[string]bool{}
		for _, link := range links {
			existing[link.AuthorityId+":"+link.MenuId] = true
		}
		for _, link := range links {
			id, err := strconv.ParseUint(link.MenuId, 10, 64)
			if err != nil {
				continue
			}
			m := byID[uint(id)]
			if m == nil || m.Hidden {
				continue
			}
			for m.ParentId != 0 {
				m = byID[m.ParentId]
				if m == nil {
					return fmt.Errorf("权限菜单父节点不存在")
				}
				menuID := strconv.FormatUint(uint64(m.ID), 10)
				key := link.AuthorityId + ":" + menuID
				if !existing[key] {
					if err := tx.Create(&model.SysAuthorityMenu{MenuId: menuID, AuthorityId: link.AuthorityId}).Error; err != nil {
						return err
					}
					existing[key] = true
				}
			}
		}
		return nil
	})
}
