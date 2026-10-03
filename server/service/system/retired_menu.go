package system

import "gorm.io/gorm"

// excludeAIPageDrawingMenu 兼容升级前的数据库，不再下发已移除页面的菜单。
// 按组件路径匹配，避免影响用户自建的同名菜单或图片字段功能。
func excludeAIPageDrawingMenu(db *gorm.DB) *gorm.DB {
	return db.Where("component IS NULL OR component <> ?", "view/systemTools/autoCode/picture.vue")
}
