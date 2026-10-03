package blog

import (
	"errors"
	"strings"
	"time"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	"gorm.io/gorm"
)

type CommentService struct{}

func (s *CommentService) GetPublicList(info blogReq.CommentSearch, rawToken string) (result map[string]interface{}, err error) {
	access, err := s.checkCommentAccess(info.Page, info.BlogID, rawToken)
	if err != nil {
		return nil, err
	}
	db := global.GVA_DB.Model(&blogModel.Comment{}).Where("page = ?", info.Page)
	if info.BlogID != nil {
		db = db.Where("blog_id = ?", *info.BlogID)
	}

	var allCount int64
	if err = db.Count(&allCount).Error; err != nil {
		return nil, err
	}

	publishedDB := db.Where("is_published = ?", true)
	var openCount int64
	if err = publishedDB.Count(&openCount).Error; err != nil {
		return nil, err
	}

	if info.PageNum <= 0 {
		info.PageNum = 1
	}
	if info.PageSize <= 0 {
		info.PageSize = 10
	}
	offset := info.PageSize * (info.PageNum - 1)
	var list []blogModel.Comment
	listDB := publishedDB
	resultTotal := openCount
	if access.IsAdmin {
		listDB = db
		resultTotal = allCount
	}
	if err = listDB.Order("create_time desc").Limit(info.PageSize).Offset(offset).Find(&list).Error; err != nil {
		return nil, err
	}

	result = map[string]interface{}{
		"allComment":   allCount,
		"closeComment": allCount - openCount,
		"comments": map[string]interface{}{
			"list":     list,
			"page":     info.PageNum,
			"pageSize": info.PageSize,
			"total":    resultTotal,
		},
	}
	return
}

func (s *CommentService) Create(info blogReq.CommentCreate) error {
	access, err := s.checkCommentAccess(info.Page, info.BlogID, info.AccessToken)
	if err != nil {
		return err
	}
	s.applyCommentDefaults(&info, access)
	now := time.Now()
	if !info.IsPublished {
		info.IsPublished = true
	}
	entity := blogModel.Comment{
		Nickname:        info.Nickname,
		Email:           info.Email,
		Content:         info.Content,
		Avatar:          info.Avatar,
		CreateTime:      &now,
		IP:              info.IP,
		IsPublished:     info.IsPublished,
		IsAdminComment:  info.IsAdminComment,
		Page:            info.Page,
		IsNotice:        info.IsNotice,
		BlogID:          info.BlogID,
		ParentCommentID: info.ParentCommentID,
		Website:         info.Website,
		QQ:              info.QQ,
	}
	return global.GVA_DB.Create(&entity).Error
}

func (s *CommentService) checkCommentAccess(page int, blogID *uint, rawToken string) (blogAccessContext, error) {
	switch page {
	case 1:
		var row blogModel.About
		if err := global.GVA_DB.Where("name_en = ?", "commentEnabled").First(&row).Error; err == nil {
			if strings.EqualFold(row.Value, "false") || row.Value == "0" {
				return blogAccessContext{}, errors.New("comment closed")
			}
		}
		return parseCommentAccess(rawToken)
	case 2:
		var row blogModel.SiteSetting
		if err := global.GVA_DB.Where("name_en = ?", "friendCommentEnabled").First(&row).Error; err == nil && row.Value != nil {
			if *row.Value == "0" || strings.EqualFold(*row.Value, "false") {
				return blogAccessContext{}, errors.New("comment closed")
			}
		}
		return parseCommentAccess(rawToken)
	default:
		if blogID == nil {
			return parseCommentAccess(rawToken)
		}
		var blog blogModel.Blog
		if err := global.GVA_DB.First(&blog, *blogID).Error; err != nil {
			return blogAccessContext{}, err
		}
		if !blog.IsCommentEnabled {
			return blogAccessContext{}, errors.New("comment closed")
		}
		return ensureBlogReadable(blog, rawToken)
	}
}

func (s *CommentService) applyCommentDefaults(info *blogReq.CommentCreate, access blogAccessContext) {
	info.Nickname = strings.TrimSpace(info.Nickname)
	info.Email = strings.TrimSpace(info.Email)
	info.Avatar = strings.TrimSpace(info.Avatar)
	if info.QQ != nil {
		qq := strings.TrimSpace(*info.QQ)
		info.QQ = &qq
		if info.Avatar == "" && isQQNumber(qq) {
			info.Avatar = "https://q1.qlogo.cn/g?b=qq&nk=" + qq + "&s=100"
		}
	}
	if info.Avatar == "" {
		info.Avatar = defaultCommentAvatar(info.Nickname)
	}
	if access.IsAdmin {
		info.IsAdminComment = true
		info.IsPublished = true
		if user, err := loadBlogAdminUser(access.Username); err == nil {
			if info.Nickname == "" {
				info.Nickname = user.NickName
			}
			if info.Email == "" {
				info.Email = user.Email
			}
			if info.Avatar == "" {
				info.Avatar = user.HeaderImg
			}
		}
	}
	if info.Nickname == "" {
		info.Nickname = "Visitor"
	}
}

func isQQNumber(value string) bool {
	if len(value) < 5 || len(value) > 12 || value[0] == '0' {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func defaultCommentAvatar(seed string) string {
	hash := uint32(0)
	for _, r := range seed {
		hash = hash*31 + uint32(r)
	}
	return "/img/comment-avatar/" + string(rune('1'+hash%6)) + ".jpg"
}

func parseCommentAccess(rawToken string) (blogAccessContext, error) {
	access, err := parseBlogAccessToken(rawToken)
	if err != nil {
		return blogAccessContext{}, nil
	}
	return access, nil
}

func (s *CommentService) GetAdminList(info blogReq.CommentAdminSearch) (list []blogModel.Comment, total int64, err error) {
	db := global.GVA_DB.Model(&blogModel.Comment{})
	if info.Page != nil {
		db = db.Where("page = ?", *info.Page)
	}
	if info.BlogID != nil {
		db = db.Where("blog_id = ?", *info.BlogID)
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	if info.PageNum <= 0 {
		info.PageNum = 1
	}
	if info.PageSize <= 0 {
		info.PageSize = 10
	}
	offset := info.PageSize * (info.PageNum - 1)
	err = db.Order("create_time desc").Limit(info.PageSize).Offset(offset).Find(&list).Error
	if err != nil {
		return
	}
	s.fillCommentBlogs(list)
	return
}

func (s *CommentService) fillCommentBlogs(list []blogModel.Comment) {
	blogIDs := make([]uint, 0, len(list))
	seen := map[uint]struct{}{}
	for _, item := range list {
		if item.BlogID == nil || *item.BlogID == 0 {
			continue
		}
		if _, ok := seen[*item.BlogID]; ok {
			continue
		}
		seen[*item.BlogID] = struct{}{}
		blogIDs = append(blogIDs, *item.BlogID)
	}
	if len(blogIDs) == 0 {
		return
	}

	var blogs []blogModel.Blog
	if err := global.GVA_DB.Select("id,title").Where("id IN ?", blogIDs).Find(&blogs).Error; err != nil {
		return
	}
	blogMap := make(map[uint]*blogModel.Blog, len(blogs))
	for i := range blogs {
		blogMap[blogs[i].ID] = &blogs[i]
	}
	for i := range list {
		if list[i].BlogID == nil {
			continue
		}
		list[i].Blog = blogMap[*list[i].BlogID]
	}
}

func (s *CommentService) UpdatePublished(id uint, published bool) error {
	return global.GVA_DB.Model(&blogModel.Comment{}).Where("id = ?", id).Update("is_published", published).Error
}

func (s *CommentService) UpdateNotice(id uint, notice bool) error {
	return global.GVA_DB.Model(&blogModel.Comment{}).Where("id = ?", id).Update("is_notice", notice).Error
}

func (s *CommentService) Delete(id uint) error {
	return global.GVA_DB.Where("id = ? OR parent_comment_id = ?", id, id).Delete(&blogModel.Comment{}).Error
}

func (s *CommentService) Update(info blogReq.CommentUpdate) error {
	return global.GVA_DB.Model(&blogModel.Comment{}).Where("id = ?", info.ID).Updates(map[string]interface{}{
		"nickname": info.Nickname,
		"email":    info.Email,
		"content":  info.Content,
		"avatar":   info.Avatar,
		"ip":       info.IP,
		"website":  info.Website,
		"qq":       info.QQ,
	}).Error
}

func (s *CommentService) GetBlogIDAndTitle() ([]blogModel.Blog, error) {
	var list []blogModel.Blog
	err := global.GVA_DB.Select("id,title").Order("create_time desc").Find(&list).Error
	return list, err
}

func (s *CommentService) GetCommentsByBlogID(id uint) ([]blogModel.Comment, error) {
	var list []blogModel.Comment
	err := global.GVA_DB.Where("blog_id = ?", id).Order("create_time desc").Find(&list).Error
	return list, err
}

func (s *CommentService) Count() (int64, error) {
	var total int64
	err := global.GVA_DB.Model(&blogModel.Comment{}).Count(&total).Error
	return total, err
}

func (s *CommentService) CountToday() (int64, error) {
	var total int64
	err := global.GVA_DB.Model(&blogModel.Comment{}).Where("date(create_time) = curdate()").Count(&total).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}
	return total, nil
}
