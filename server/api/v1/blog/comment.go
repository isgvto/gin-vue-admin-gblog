package blog

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"github.com/isgvto/gin-vue-admin-gblog/server/utils"
	"go.uber.org/zap"
)

type CommentApi struct{}

func (a *CommentApi) GetComments(c *gin.Context) {
	var req blogReq.CommentSearch
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	data, err := commentService.GetPublicList(req, c.GetHeader("Authorization"))
	if err != nil {
		global.GVA_LOG.Error("get comments failed", zap.Error(err))
		response.FailWithMessage("获取评论列表失败", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *CommentApi) CreateComment(c *gin.Context) {
	var req blogReq.CommentCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	req.AccessToken = c.GetHeader("Authorization")
	if req.IP == nil {
		ip := utils.BlogClientIP(c.Request)
		req.IP = &ip
	}
	if err := commentService.Create(req); err != nil {
		global.GVA_LOG.Error("create comment failed", zap.Error(err))
		response.FailWithMessage("创建评论失败", c)
		return
	}
	response.OkWithMessage("创建评论成功", c)
}

func (a *CommentApi) GetCommentList(c *gin.Context) {
	var req blogReq.CommentAdminSearch
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := commentService.GetAdminList(req)
	if err != nil {
		global.GVA_LOG.Error("get admin comments failed", zap.Error(err))
		response.FailWithMessage("获取评论列表失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.PageNum, PageSize: req.PageSize}, "获取评论列表成功", c)
}

func (a *CommentApi) GetBlogIDAndTitle(c *gin.Context) {
	list, err := commentService.GetBlogIDAndTitle()
	if err != nil {
		global.GVA_LOG.Error("get blog id and title failed", zap.Error(err))
		response.FailWithMessage("获取博客列表失败", c)
		return
	}
	response.OkWithData(list, c)
}

func (a *CommentApi) UpdateCommentPublished(c *gin.Context) {
	var req blogReq.ToggleReq
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := commentService.UpdatePublished(req.ID, req.Value); err != nil {
		global.GVA_LOG.Error("update comment published failed", zap.Error(err))
		response.FailWithMessage("更新评论公开状态失败", c)
		return
	}
	response.OkWithMessage("更新评论公开状态成功", c)
}

func (a *CommentApi) UpdateCommentNotice(c *gin.Context) {
	var req blogReq.ToggleReq
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := commentService.UpdateNotice(req.ID, req.Value); err != nil {
		global.GVA_LOG.Error("update comment notice failed", zap.Error(err))
		response.FailWithMessage("更新评论通知状态失败", c)
		return
	}
	response.OkWithMessage("更新评论通知状态成功", c)
}

func (a *CommentApi) UpdateComment(c *gin.Context) {
	var req blogReq.CommentUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := commentService.Update(req); err != nil {
		global.GVA_LOG.Error("update comment failed", zap.Error(err))
		response.FailWithMessage("更新评论失败", c)
		return
	}
	response.OkWithMessage("更新评论成功", c)
}

func (a *CommentApi) DeleteComment(c *gin.Context) {
	var req blogReq.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := commentService.Delete(req.ID); err != nil {
		global.GVA_LOG.Error("delete comment failed", zap.Error(err))
		response.FailWithMessage("删除评论失败", c)
		return
	}
	response.OkWithMessage("删除评论成功", c)
}
