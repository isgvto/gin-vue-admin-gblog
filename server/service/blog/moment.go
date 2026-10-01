package blog

import (
	"time"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/utils"
	"gorm.io/gorm"
)

type MomentService struct{}

func (s *MomentService) GetPublicList(info blogReq.PageQuery) (list []blogModel.Moment, total int64, err error) {
	db := global.GVA_DB.Model(&blogModel.Moment{}).Where("is_published = ?", true)
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	if info.Page <= 0 {
		info.Page = 1
	}
	if info.PageSize <= 0 {
		info.PageSize = 10
	}
	offset := info.PageSize * (info.Page - 1)
	err = db.Order("create_time desc").Limit(info.PageSize).Offset(offset).Find(&list).Error
	if err == nil {
		renderMomentListContent(list)
	}
	return
}

func (s *MomentService) AddLike(id uint) error {
	return global.GVA_DB.Model(&blogModel.Moment{}).Where("id = ?", id).UpdateColumn("likes", gorm.Expr("coalesce(likes,0) + 1")).Error
}

func (s *MomentService) GetList(info blogReq.PageQuery) (list []blogModel.Moment, total int64, err error) {
	db := global.GVA_DB.Model(&blogModel.Moment{})
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	if info.Page <= 0 {
		info.Page = 1
	}
	if info.PageSize <= 0 {
		info.PageSize = 10
	}
	offset := info.PageSize * (info.Page - 1)
	err = db.Order("create_time desc").Limit(info.PageSize).Offset(offset).Find(&list).Error
	return
}

func (s *MomentService) GetByID(id uint) (moment blogModel.Moment, err error) {
	err = global.GVA_DB.First(&moment, id).Error
	return
}

func (s *MomentService) Create(info blogReq.MomentUpsert) error {
	createTime := time.Now()
	if info.CreateTime != nil {
		createTime = *info.CreateTime
	}
	entity := blogModel.Moment{
		Content:     info.Content,
		CreateTime:  createTime,
		Likes:       info.Likes,
		IsPublished: info.IsPublished,
	}
	return global.GVA_DB.Create(&entity).Error
}

func (s *MomentService) Update(info blogReq.MomentUpsert) error {
	updates := map[string]interface{}{
		"content":      info.Content,
		"likes":        info.Likes,
		"is_published": info.IsPublished,
	}
	if info.CreateTime != nil {
		updates["create_time"] = *info.CreateTime
	}
	return global.GVA_DB.Model(&blogModel.Moment{}).Where("id = ?", info.ID).Updates(updates).Error
}

func (s *MomentService) Delete(id uint) error {
	return global.GVA_DB.Delete(&blogModel.Moment{}, id).Error
}

func (s *MomentService) UpdatePublished(id uint, published bool) error {
	return global.GVA_DB.Model(&blogModel.Moment{}).Where("id = ?", id).Update("is_published", published).Error
}

func renderMomentListContent(list []blogModel.Moment) {
	for i := range list {
		list[i].Content = utils.MarkdownToHTML(list[i].Content)
	}
}
