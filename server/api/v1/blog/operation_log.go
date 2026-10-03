package blog

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"go.uber.org/zap"
)

type OperationLogApi struct{}

func (a *OperationLogApi) GetOperationLogs(c *gin.Context) {
	var req blogReq.DateRangePageQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := operationLogService.GetList(req)
	if err != nil {
		global.GVA_LOG.Error("get operation logs failed", zap.Error(err))
		response.FailWithMessage("获取操作日志失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "获取操作日志成功", c)
}

func (a *OperationLogApi) DeleteOperationLog(c *gin.Context) {
	var req blogReq.IDReq
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := operationLogService.Delete(req.ID); err != nil {
		global.GVA_LOG.Error("delete operation log failed", zap.Error(err))
		response.FailWithMessage("删除操作日志失败", c)
		return
	}
	response.OkWithMessage("删除操作日志成功", c)
}
