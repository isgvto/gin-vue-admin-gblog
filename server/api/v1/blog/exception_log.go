package blog

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"go.uber.org/zap"
)

type ExceptionLogApi struct{}

func (a *ExceptionLogApi) GetExceptionLogs(c *gin.Context) {
	var req blogReq.DateRangePageQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := exceptionLogService.GetList(req)
	if err != nil {
		global.GVA_LOG.Error("get exception logs failed", zap.Error(err))
		response.FailWithMessage("获取异常日志失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "获取异常日志成功", c)
}

func (a *ExceptionLogApi) DeleteExceptionLog(c *gin.Context) {
	var req blogReq.IDReq
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := exceptionLogService.Delete(req.ID); err != nil {
		global.GVA_LOG.Error("delete exception log failed", zap.Error(err))
		response.FailWithMessage("删除异常日志失败", c)
		return
	}
	response.OkWithMessage("删除异常日志成功", c)
}
