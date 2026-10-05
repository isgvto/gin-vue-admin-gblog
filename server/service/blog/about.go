package blog

import (
	"errors"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	"github.com/isgvto/gin-vue-admin-gblog/server/utils"
)

type AboutService struct{}

func (s *AboutService) GetList() ([]blogModel.About, error) {
	var list []blogModel.About
	err := global.GVA_DB.Where("name_en <> ?", "musicId").Order("id asc").Find(&list).Error
	return list, err
}

func (s *AboutService) GetPublicList() ([]blogModel.About, error) {
	list, err := s.GetList()
	if err != nil {
		return nil, err
	}
	for i := range list {
		if list[i].NameEn == "content" {
			list[i].Value = utils.MarkdownToHTML(list[i].Value)
		}
	}
	return list, nil
}

func (s *AboutService) UpdateValues(values map[string]string) error {
	if len(values) == 0 {
		return errors.New("about values is empty")
	}

	tx := global.GVA_DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for key, value := range values {
		if key == "musicId" {
			continue
		}
		result := tx.Model(&blogModel.About{}).Where("name_en = ?", key).Update("value", value)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
		if result.RowsAffected == 0 {
			if err := tx.Create(&blogModel.About{
				NameEn: key,
				NameZh: aboutNameZh(key),
				Value:  value,
			}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	return tx.Commit().Error
}

func aboutNameZh(nameEn string) string {
	switch nameEn {
	case "title":
		return "标题"
	case "content":
		return "正文Markdown"
	case "commentEnabled":
		return "评论开关"
	default:
		return nameEn
	}
}
