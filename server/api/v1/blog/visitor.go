package blog

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	commonReq "github.com/isgvto/gin-vue-admin-gblog/server/model/common/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"go.uber.org/zap"
)

type VisitorApi struct{}

func (a *VisitorApi) GetVisitors(c *gin.Context) {
	var req blogReq.DateRangePageQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := visitorService.GetList(req)
	if err != nil {
		global.GVA_LOG.Error("get visitors failed", zap.Error(err))
		response.FailWithMessage("获取访客列表失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "获取访客列表成功", c)
}

func (a *VisitorApi) DeleteVisitor(c *gin.Context) {
	var req blogReq.IDReq
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := visitorService.Delete(req.ID); err != nil {
		global.GVA_LOG.Error("delete visitor failed", zap.Error(err))
		response.FailWithMessage("删除访客失败", c)
		return
	}
	response.OkWithMessage("删除访客成功", c)
}

func (a *VisitorApi) DeleteVisitors(c *gin.Context) {
	var req commonReq.IdsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if len(req.Ids) == 0 {
		response.FailWithMessage("请选择要删除的访客", c)
		return
	}
	if err := visitorService.DeleteByIds(req.Ids); err != nil {
		global.GVA_LOG.Error("delete visitors failed", zap.Error(err))
		response.FailWithMessage("批量删除访客失败", c)
		return
	}
	response.OkWithMessage("批量删除访客成功", c)
}
