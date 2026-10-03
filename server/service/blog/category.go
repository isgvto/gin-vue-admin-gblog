package blog

import (
	"errors"
	"strings"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	blogResp "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/response"
)

type CategoryService struct{}

func (s *CategoryService) GetList() ([]blogModel.Category, error) {
	var list []blogModel.Category
	err := global.GVA_DB.Order("id asc").Find(&list).Error
	return list, err
}

func (s *CategoryService) GetPublishedBlogListByName(info blogReq.CategoryBlogSearch) (list []blogResp.BlogInfoItem, total int64, err error) {
	categoryName := strings.TrimSpace(info.CategoryName)
	if categoryName == "" {
		return nil, 0, errors.New("category name is required")
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
		Joins("JOIN gvto_category ON gvto_category.id = gvto_blog.category_id").
		Where("gvto_blog.is_published = ? AND gvto_category.category_name = ?", true, categoryName)
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

func (s *CategoryService) Create(info blogReq.CategoryUpsert) error {
	name := strings.TrimSpace(info.CategoryName)
	if name == "" {
		return errors.New("category name is required")
	}
	var existing blogModel.Category
	if err := global.GVA_DB.Where("category_name = ?", name).First(&existing).Error; err == nil {
		return errors.New("category already exists")
	}
	entity := blogModel.Category{CategoryName: name}
	return global.GVA_DB.Create(&entity).Error
}

func (s *CategoryService) Update(info blogReq.CategoryUpsert) error {
	name := strings.TrimSpace(info.CategoryName)
	if name == "" {
		return errors.New("category name is required")
	}
	var existing blogModel.Category
	if err := global.GVA_DB.Where("category_name = ?", name).First(&existing).Error; err == nil && existing.ID != info.ID {
		return errors.New("category already exists")
	}
	return global.GVA_DB.Model(&blogModel.Category{}).
		Where("id = ?", info.ID).
		Update("category_name", name).Error
}

func (s *CategoryService) Delete(id uint) error {
	var count int64
	if err := global.GVA_DB.Model(&blogModel.Blog{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("category is referenced by blogs")
	}
	return global.GVA_DB.Delete(&blogModel.Category{}, id).Error
}
