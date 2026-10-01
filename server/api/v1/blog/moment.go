package blog

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"go.uber.org/zap"
)

type MomentApi struct{}

func (a *MomentApi) GetMoments(c *gin.Context) {
	var req blogReq.PageQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := momentService.GetPublicList(req)
	if err != nil {
		global.GVA_LOG.Error("get moments failed", zap.Error(err))
		response.FailWithMessage("获取动态列表失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "获取动态列表成功", c)
}

func (a *MomentApi) LikeMoment(c *gin.Context) {
	var req blogReq.IDUriReq
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := momentService.AddLike(req.ID); err != nil {
		global.GVA_LOG.Error("like moment failed", zap.Error(err))
		response.FailWithMessage("点赞失败", c)
		return
	}
	response.OkWithMessage("点赞成功", c)
}

func (a *MomentApi) GetMomentList(c *gin.Context) {
	var req blogReq.PageQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := momentService.GetList(req)
	if err != nil {
		global.GVA_LOG.Error("get admin moments failed", zap.Error(err))
		response.FailWithMessage("获取动态列表失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "获取动态列表成功", c)
}

func (a *MomentApi) GetMoment(c *gin.Context) {
	var req blogReq.IDQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	data, err := momentService.GetByID(req.ID)
	if err != nil {
		global.GVA_LOG.Error("get moment failed", zap.Error(err))
		response.FailWithMessage("获取动态失败", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *MomentApi) CreateMoment(c *gin.Context) {
	var req blogReq.MomentUpsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := momentService.Create(req); err != nil {
		global.GVA_LOG.Error("create moment failed", zap.Error(err))
		response.FailWithMessage("创建动态失败", c)
		return
	}
	response.OkWithMessage("创建动态成功", c)
}

func (a *MomentApi) UpdateMoment(c *gin.Context) {
	var req blogReq.MomentUpsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := momentService.Update(req); err != nil {
		global.GVA_LOG.Error("update moment failed", zap.Error(err))
		response.FailWithMessage("更新动态失败", c)
		return
	}
	response.OkWithMessage("更新动态成功", c)
}

func (a *MomentApi) DeleteMoment(c *gin.Context) {
	var req blogReq.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := momentService.Delete(req.ID); err != nil {
		global.GVA_LOG.Error("delete moment failed", zap.Error(err))
		response.FailWithMessage("删除动态失败", c)
		return
	}
	response.OkWithMessage("删除动态成功", c)
}

func (a *MomentApi) UpdateMomentPublished(c *gin.Context) {
	var req blogReq.ToggleReq
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := momentService.UpdatePublished(req.ID, req.Value); err != nil {
		global.GVA_LOG.Error("update moment published failed", zap.Error(err))
		response.FailWithMessage("更新动态公开状态失败", c)
		return
	}
	response.OkWithMessage("更新动态公开状态成功", c)
}
