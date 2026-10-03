package blog

import (
	"errors"
	"strings"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	blogResp "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/response"
)

type TagService struct{}

func (s *TagService) GetList() ([]blogModel.Tag, error) {
	var list []blogModel.Tag
	err := global.GVA_DB.Order("id asc").Find(&list).Error
	return list, err
}

func (s *TagService) GetPublishedBlogListByName(info blogReq.TagBlogSearch) (list []blogResp.BlogInfoItem, total int64, err error) {
	tagName := strings.TrimSpace(info.TagName)
	if tagName == "" {
		return nil, 0, errors.New("tag name is required")
	}
	if info.Page <= 0 {
		info.Page = info.PageNum
	}
	if info.Page <= 0 {
		info.Page = 1
	}
	if info.PageSize <= 0 {
		info.PageSize = 10
	}

	db := global.GVA_DB.Model(&blogModel.Blog{}).
		Joins("JOIN gvto_blog_tag ON gvto_blog_tag.blog_id = gvto_blog.id").
		Joins("JOIN gvto_tag ON gvto_tag.id = gvto_blog_tag.tag_id").
		Where("gvto_blog.is_published = ? AND gvto_tag.tag_name = ?", true, tagName)
	if err = db.Count(&total).Error; err != nil {
		return
	}

	var blogs []blogModel.Blog
	offset := info.PageSize * (info.Page - 1)
	err = db.Order("gvto_blog.is_top desc, gvto_blog.create_time desc").
		Limit(info.PageSize).
		Offset(offset).
		Find(&blogs).Error
	if err != nil {
		return
	}
	return buildBlogInfoItems(blogs), total, nil
}

func (s *TagService) Create(info blogReq.TagUpsert) error {
	name := strings.TrimSpace(info.TagName)
	if name == "" {
		return errors.New("tag name is required")
	}
	var existing blogModel.Tag
	if err := global.GVA_DB.Where("tag_name = ?", name).First(&existing).Error; err == nil {
		return errors.New("tag already exists")
	}
	entity := blogModel.Tag{TagName: name, Color: info.Color}
	return global.GVA_DB.Create(&entity).Error
}

func (s *TagService) Update(info blogReq.TagUpsert) error {
	name := strings.TrimSpace(info.TagName)
	if name == "" {
		return errors.New("tag name is required")
	}
	var existing blogModel.Tag
	if err := global.GVA_DB.Where("tag_name = ?", name).First(&existing).Error; err == nil && existing.ID != info.ID {
		return errors.New("tag already exists")
	}
	updates := map[string]interface{}{
		"tag_name": name,
		"color":    info.Color,
	}
	return global.GVA_DB.Model(&blogModel.Tag{}).
		Where("id = ?", info.ID).
		Updates(updates).Error
}

func (s *TagService) Delete(id uint) error {
	var count int64
	if err := global.GVA_DB.Model(&blogModel.BlogTag{}).Where("tag_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("tag is referenced by blogs")
	}
	return global.GVA_DB.Delete(&blogModel.Tag{}, id).Error
}
