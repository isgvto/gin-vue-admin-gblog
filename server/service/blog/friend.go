package blog

import (
	"time"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	blogResp "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/response"
	"github.com/isgvto/gin-vue-admin-gblog/server/utils"
	"gorm.io/gorm"
)

type FriendService struct{}

func (s *FriendService) GetPublicPage() (map[string]interface{}, error) {
	var friendList []blogModel.Friend
	if err := global.GVA_DB.Where("is_published = ?", true).Order("create_time asc").Find(&friendList).Error; err != nil {
		return nil, err
	}
	info, err := s.GetInfo()
	if err != nil {
		return nil, err
	}
	info.Content = utils.MarkdownToHTML(info.Content)
	return map[string]interface{}{
		"friendList": friendList,
		"friendInfo": info,
	}, nil
}

func (s *FriendService) AddViewsByNickname(nickname string) error {
	return global.GVA_DB.Model(&blogModel.Friend{}).
		Where("nickname = ?", nickname).
		UpdateColumn("views", gorm.Expr("views + ?", 1)).Error
}

func (s *FriendService) GetList(info blogReq.PageQuery) (list []blogModel.Friend, total int64, err error) {
	db := global.GVA_DB.Model(&blogModel.Friend{})
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
	err = db.Order("create_time asc").Limit(info.PageSize).Offset(offset).Find(&list).Error
	return
}

func (s *FriendService) Create(info blogReq.FriendUpsert) error {
	entity := blogModel.Friend{
		Nickname:    info.Nickname,
		Description: info.Description,
		Website:     info.Website,
		Avatar:      info.Avatar,
		IsPublished: info.IsPublished,
		Views:       info.Views,
		CreateTime:  time.Now(),
	}
	return global.GVA_DB.Create(&entity).Error
}

func (s *FriendService) Update(info blogReq.FriendUpsert) error {
	return global.GVA_DB.Model(&blogModel.Friend{}).Where("id = ?", info.ID).Updates(map[string]interface{}{
		"nickname":     info.Nickname,
		"description":  info.Description,
		"website":      info.Website,
		"avatar":       info.Avatar,
		"is_published": info.IsPublished,
		"views":        info.Views,
	}).Error
}

func (s *FriendService) Delete(id uint) error {
	return global.GVA_DB.Delete(&blogModel.Friend{}, id).Error
}

func (s *FriendService) UpdatePublished(id uint, published bool) error {
	return global.GVA_DB.Model(&blogModel.Friend{}).Where("id = ?", id).Update("is_published", published).Error
}

func (s *FriendService) GetInfo() (blogResp.FriendInfoResponse, error) {
	var contentRow blogModel.SiteSetting
	var commentRow blogModel.SiteSetting
	if err := global.GVA_DB.Where("name_en = ?", "friendContent").First(&contentRow).Error; err != nil && err != gorm.ErrRecordNotFound {
		return blogResp.FriendInfoResponse{}, err
	}
	if err := global.GVA_DB.Where("name_en = ?", "friendCommentEnabled").First(&commentRow).Error; err != nil && err != gorm.ErrRecordNotFound {
		return blogResp.FriendInfoResponse{}, err
	}
	resp := blogResp.FriendInfoResponse{}
	if contentRow.Value != nil {
		resp.Content = *contentRow.Value
	}
	if commentRow.Value != nil {
		resp.CommentEnabled = *commentRow.Value == "1" || *commentRow.Value == "true"
	}
	return resp, nil
}

func (s *FriendService) UpdateInfoContent(content string) error {
	return global.GVA_DB.Model(&blogModel.SiteSetting{}).Where("name_en = ?", "friendContent").Update("value", content).Error
}

func (s *FriendService) UpdateInfoCommentEnabled(enabled bool) error {
	value := "0"
	if enabled {
		value = "1"
	}
	return global.GVA_DB.Model(&blogModel.SiteSetting{}).Where("name_en = ?", "friendCommentEnabled").Update("value", value).Error
}
