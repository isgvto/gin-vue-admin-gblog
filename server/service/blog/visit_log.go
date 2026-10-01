package blog

import (
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
)

type VisitLogService struct{}

func (s *VisitLogService) GetList(info blogReq.DateRangePageQuery, uuid string) (list []blogModel.VisitLog, total int64, err error) {
	db := global.GVA_DB.Model(&blogModel.VisitLog{})
	if uuid != "" {
		db = db.Where("uuid LIKE ?", "%"+uuid+"%")
	}
	if info.StartDate != "" && info.EndDate != "" {
		db = db.Where("create_time BETWEEN ? AND ?", info.StartDate, info.EndDate)
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

func (s *VisitLogService) Delete(id uint) error {
	return global.GVA_DB.Delete(&blogModel.VisitLog{}, id).Error
}

func (s *VisitLogService) DeleteByIds(ids []int) error {
	return global.GVA_DB.Delete(&[]blogModel.VisitLog{}, "id in ?", ids).Error
}
