package blog

import (
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
)

type OperationLogCompatService struct{}

func (s *OperationLogCompatService) GetList(info blogReq.DateRangePageQuery) (list []blogModel.OperationLog, total int64, err error) {
	db := global.GVA_DB.Model(&blogModel.OperationLog{})
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

func (s *OperationLogCompatService) Delete(id uint) error {
	return global.GVA_DB.Delete(&blogModel.OperationLog{}, id).Error
}
