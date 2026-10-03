package blog

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	commonReq "github.com/isgvto/gin-vue-admin-gblog/server/model/common/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"go.uber.org/zap"
)

type VisitLogApi struct{}

func (a *VisitLogApi) GetVisitLogs(c *gin.Context) {
	var req blogReq.DateRangePageQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	uuid := c.Query("uuid")
	list, total, err := visitLogService.GetList(req, uuid)
	if err != nil {
		global.GVA_LOG.Error("get visit logs failed", zap.Error(err))
		response.FailWithMessage("获取访问日志失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "获取访问日志成功", c)
}

func (a *VisitLogApi) DeleteVisitLog(c *gin.Context) {
	var req blogReq.IDReq
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := visitLogService.Delete(req.ID); err != nil {
		global.GVA_LOG.Error("delete visit log failed", zap.Error(err))
		response.FailWithMessage("删除访问日志失败", c)
		return
	}
	response.OkWithMessage("删除访问日志成功", c)
}

func (a *VisitLogApi) DeleteVisitLogs(c *gin.Context) {
	var req commonReq.IdsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if len(req.Ids) == 0 {
		response.FailWithMessage("请选择要删除的访问日志", c)
		return
	}
	if err := visitLogService.DeleteByIds(req.Ids); err != nil {
		global.GVA_LOG.Error("delete visit logs failed", zap.Error(err))
		response.FailWithMessage("批量删除访问日志失败", c)
		return
	}
	response.OkWithMessage("批量删除访问日志成功", c)
}
