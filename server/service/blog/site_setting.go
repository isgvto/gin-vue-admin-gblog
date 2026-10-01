package blog

import (
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	"gorm.io/gorm"
)

type SiteSettingService struct{}

type defaultSiteSetting struct {
	NameEn string
	NameZh string
	Value  string
	Type   int
}

var defaultSiteSettings = []defaultSiteSetting{
	{NameEn: "bg1", NameZh: "首页背景图 1", Value: "https://www.guitu.life/blog-file/bg1.jpg", Type: 1},
	{NameEn: "bg2", NameZh: "首页背景图 2", Value: "https://www.guitu.life/blog-file/bg2.jpg", Type: 1},
	{NameEn: "bg3", NameZh: "首页背景图 3", Value: "https://www.guitu.life/blog-file/bg3.jpg", Type: 1},
	{NameEn: "malfunctionText", NameZh: "首页故障风文字", Value: "Gvto's Blog", Type: 1},
	{NameEn: "docsGithubRepo", NameZh: "GitHub文档仓库", Value: "", Type: 4},
	{NameEn: "docsGithubBranch", NameZh: "文档仓库分支", Value: "main", Type: 4},
	{NameEn: "docsGithubRoot", NameZh: "文档根目录", Value: "", Type: 4},
	{NameEn: "docsGithubWebhookSecret", NameZh: "Webhook密钥", Value: "", Type: 4},
}

func (s *SiteSettingService) GetGrouped() (map[string][]blogModel.SiteSetting, error) {
	if err := ensureDefaultSiteSettings(); err != nil {
		return nil, err
	}

	var list []blogModel.SiteSetting
	if err := global.GVA_DB.Order("id asc").Find(&list).Error; err != nil {
		return nil, err
	}
	result := map[string][]blogModel.SiteSetting{}
	for _, item := range list {
		key := "type0"
		if item.Type != nil {
			key = "type" + string(rune('0'+*item.Type))
		}
		result[key] = append(result[key], item)
	}
	return result, nil
}

func ensureDefaultSiteSettings() error {
	for _, item := range defaultSiteSettings {
		var existing blogModel.SiteSetting
		err := global.GVA_DB.Where("name_en = ?", item.NameEn).First(&existing).Error
		if err == nil {
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}

		nameEn := item.NameEn
		nameZh := item.NameZh
		value := item.Value
		settingType := item.Type
		if err := global.GVA_DB.Create(&blogModel.SiteSetting{
			NameEn: &nameEn,
			NameZh: &nameZh,
			Value:  &value,
			Type:   &settingType,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}

func (s *SiteSettingService) UpdateAll(info blogReq.SiteSettingBatchUpdate) error {
	tx := global.GVA_DB.Begin()
	for _, id := range info.DeleteIDs {
		if err := tx.Delete(&blogModel.SiteSetting{}, id).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	for _, item := range info.Settings {
		if item.ID > 0 {
			if err := tx.Model(&blogModel.SiteSetting{}).Where("id = ?", item.ID).Updates(map[string]interface{}{
				"name_en": item.NameEn,
				"name_zh": item.NameZh,
				"value":   item.Value,
				"type":    item.Type,
			}).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			entity := blogModel.SiteSetting{
				NameEn: item.NameEn,
				NameZh: item.NameZh,
				Value:  item.Value,
				Type:   item.Type,
			}
			if err := tx.Create(&entity).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	return tx.Commit().Error
}

func (s *SiteSettingService) GetWebTitleSuffix() (string, error) {
	var row blogModel.SiteSetting
	err := global.GVA_DB.Where("name_en = ?", "webTitleSuffix").First(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	if row.Value == nil {
		return "", nil
	}
	return *row.Value, nil
}
