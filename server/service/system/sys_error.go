package system

import (
	"context"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	systemReq "github.com/isgvto/gin-vue-admin-gblog/server/model/system/request"
)

type SysErrorService struct{}

// CreateSysError 创建错误日志记录
// Author [yourname](https://github.com/yourname)
func (sysErrorService *SysErrorService) CreateSysError(ctx context.Context, sysError *system.SysError) (err error) {
	if global.GVA_DB == nil {
		return nil
	}
	// Public log ingestion cannot set server-owned analysis state.
	sysError.Status = "未处理"
	sysError.Solution = nil
	sysError.AnalysisTask = ""
	sysError.AnalysisError = ""
	sysError.AnalysisModel = ""
	sysError.AnalysisModelID = 0
	sysError.AnalysisStartedAt = nil
	sysError.AnalysisCompletedAt = nil
	sysError.SolutionModel = ""
	sysError.SolutionGeneratedAt = nil
	err = global.GVA_DB.Create(sysError).Error
	return err
}

// DeleteSysError 删除错误日志记录
// Author [yourname](https://github.com/yourname)
func (sysErrorService *SysErrorService) DeleteSysError(ctx context.Context, ID string) (err error) {
	err = global.GVA_DB.Delete(&system.SysError{}, "id = ?", ID).Error
	return err
}

// DeleteSysErrorByIds 批量删除错误日志记录
// Author [yourname](https://github.com/yourname)
func (sysErrorService *SysErrorService) DeleteSysErrorByIds(ctx context.Context, IDs []string) (err error) {
	err = global.GVA_DB.Delete(&[]system.SysError{}, "id in ?", IDs).Error
	return err
}

// UpdateSysError 更新错误日志记录
// Author [yourname](https://github.com/yourname)
func (sysErrorService *SysErrorService) UpdateSysError(ctx context.Context, sysError system.SysError) (err error) {
	err = global.GVA_DB.Model(&system.SysError{}).Where("id = ?", sysError.ID).Select("form", "info", "level").Updates(&sysError).Error
	return err
}

// GetSysError 根据ID获取错误日志记录
// Author [yourname](https://github.com/yourname)
func (sysErrorService *SysErrorService) GetSysError(ctx context.Context, ID string) (sysError system.SysError, err error) {
	if err = recoverExpiredErrorAnalysis(ctx); err != nil {
		return
	}
	err = global.GVA_DB.Where("id = ?", ID).First(&sysError).Error
	return
}

// GetSysErrorInfoList 分页获取错误日志记录
// Author [yourname](https://github.com/yourname)
func (sysErrorService *SysErrorService) GetSysErrorInfoList(ctx context.Context, info systemReq.SysErrorSearch) (list []system.SysError, total int64, err error) {
	if err = recoverExpiredErrorAnalysis(ctx); err != nil {
		return
	}
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// 创建db
	db := global.GVA_DB.Model(&system.SysError{}).Order("created_at desc")
	var sysErrors []system.SysError
	// 如果有条件搜索 下方会自动创建搜索语句
	if len(info.CreatedAtRange) == 2 {
		db = db.Where("created_at BETWEEN ? AND ?", info.CreatedAtRange[0], info.CreatedAtRange[1])
	}

	if info.Form != nil && *info.Form != "" {
		db = db.Where("form = ?", *info.Form)
	}
	if info.Info != nil && *info.Info != "" {
		db = db.Where("info LIKE ?", "%"+*info.Info+"%")
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}

	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}

	err = db.Find(&sysErrors).Error
	return sysErrors, total, err
}
