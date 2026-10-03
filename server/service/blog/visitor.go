package blog

import (
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
)

type VisitorService struct{}

func (s *VisitorService) GetList(info blogReq.DateRangePageQuery) (list []blogModel.Visitor, total int64, err error) {
	db := global.GVA_DB.Model(&blogModel.Visitor{})
	if info.StartDate != "" && info.EndDate != "" {
		db = db.Where("last_time BETWEEN ? AND ?", info.StartDate, info.EndDate)
	}
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

func (s *VisitorService) Delete(id uint) error {
	return global.GVA_DB.Delete(&blogModel.Visitor{}, id).Error
}

func (s *VisitorService) DeleteByIds(ids []int) error {
	return global.GVA_DB.Delete(&[]blogModel.Visitor{}, "id in ?", ids).Error
}
