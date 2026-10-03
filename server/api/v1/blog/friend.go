package blog

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"go.uber.org/zap"
)

type FriendApi struct{}

func (a *FriendApi) GetFriends(c *gin.Context) {
	data, err := friendService.GetPublicPage()
	if err != nil {
		global.GVA_LOG.Error("get friends failed", zap.Error(err))
		response.FailWithMessage("获取友链页失败", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *FriendApi) AddFriendViews(c *gin.Context) {
	var req blogReq.NicknameReq
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := friendService.AddViewsByNickname(req.Nickname); err != nil {
		global.GVA_LOG.Error("add friend views failed", zap.Error(err))
		response.FailWithMessage("更新友链点击数失败", c)
		return
	}
	response.OkWithMessage("更新友链点击数成功", c)
}

func (a *FriendApi) GetFriendList(c *gin.Context) {
	var req blogReq.PageQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := friendService.GetList(req)
	if err != nil {
		global.GVA_LOG.Error("get admin friends failed", zap.Error(err))
		response.FailWithMessage("获取友链列表失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: req.Page, PageSize: req.PageSize}, "获取友链列表成功", c)
}

func (a *FriendApi) CreateFriend(c *gin.Context) {
	var req blogReq.FriendUpsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := friendService.Create(req); err != nil {
		global.GVA_LOG.Error("create friend failed", zap.Error(err))
		response.FailWithMessage("创建友链失败", c)
		return
	}
	response.OkWithMessage("创建友链成功", c)
}

func (a *FriendApi) UpdateFriend(c *gin.Context) {
	var req blogReq.FriendUpsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := friendService.Update(req); err != nil {
		global.GVA_LOG.Error("update friend failed", zap.Error(err))
		response.FailWithMessage("更新友链失败", c)
		return
	}
	response.OkWithMessage("更新友链成功", c)
}

func (a *FriendApi) DeleteFriend(c *gin.Context) {
	var req blogReq.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := friendService.Delete(req.ID); err != nil {
		global.GVA_LOG.Error("delete friend failed", zap.Error(err))
		response.FailWithMessage("删除友链失败", c)
		return
	}
	response.OkWithMessage("删除友链成功", c)
}

func (a *FriendApi) UpdateFriendPublished(c *gin.Context) {
	var req blogReq.ToggleReq
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := friendService.UpdatePublished(req.ID, req.Value); err != nil {
		global.GVA_LOG.Error("update friend published failed", zap.Error(err))
		response.FailWithMessage("更新友链公开状态失败", c)
		return
	}
	response.OkWithMessage("更新友链公开状态成功", c)
}

func (a *FriendApi) GetFriendInfo(c *gin.Context) {
	data, err := friendService.GetInfo()
	if err != nil {
		global.GVA_LOG.Error("get friend info failed", zap.Error(err))
		response.FailWithMessage("获取友链页信息失败", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *FriendApi) UpdateFriendInfoContent(c *gin.Context) {
	var req blogReq.ContentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := friendService.UpdateInfoContent(req.Content); err != nil {
		global.GVA_LOG.Error("update friend info content failed", zap.Error(err))
		response.FailWithMessage("更新友链页内容失败", c)
		return
	}
	response.OkWithMessage("更新友链页内容成功", c)
}

func (a *FriendApi) UpdateFriendInfoCommentEnabled(c *gin.Context) {
	var req blogReq.ToggleReq
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := friendService.UpdateInfoCommentEnabled(req.Value); err != nil {
		global.GVA_LOG.Error("update friend comment enabled failed", zap.Error(err))
		response.FailWithMessage("更新友链评论开关失败", c)
		return
	}
	response.OkWithMessage("更新友链评论开关成功", c)
}
