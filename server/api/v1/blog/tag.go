package blog

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"go.uber.org/zap"
)

type TagApi struct{}

func (a *TagApi) GetTagList(c *gin.Context) {
	list, err := tagService.GetList()
	if err != nil {
		global.GVA_LOG.Error("get tag list failed", zap.Error(err))
		response.FailWithMessage("获取标签列表失败", c)
		return
	}
	response.OkWithData(list, c)
}

func (a *TagApi) GetTagBlogList(c *gin.Context) {
	var req blogReq.TagBlogSearch
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := tagService.GetPublishedBlogListByName(req)
	if err != nil {
		global.GVA_LOG.Error("get tag blog list failed", zap.Error(err))
		response.FailWithMessage("鑾峰彇鏍囩鏂囩珷鍒楄〃澶辫触", c)
		return
	}
	if req.Page <= 0 {
		req.Page = req.PageNum
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "鑾峰彇鏍囩鏂囩珷鍒楄〃鎴愬姛", c)
}

func (a *TagApi) CreateTag(c *gin.Context) {
	var req blogReq.TagUpsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := tagService.Create(req); err != nil {
		global.GVA_LOG.Error("create tag failed", zap.Error(err))
		response.FailWithMessage("创建标签失败", c)
		return
	}
	response.OkWithMessage("创建标签成功", c)
}

func (a *TagApi) UpdateTag(c *gin.Context) {
	var req blogReq.TagUpsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := tagService.Update(req); err != nil {
		global.GVA_LOG.Error("update tag failed", zap.Error(err))
		response.FailWithMessage("更新标签失败", c)
		return
	}
	response.OkWithMessage("更新标签成功", c)
}

func (a *TagApi) DeleteTag(c *gin.Context) {
	var req blogReq.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := tagService.Delete(req.ID); err != nil {
		global.GVA_LOG.Error("delete tag failed", zap.Error(err))
		response.FailWithMessage("删除标签失败", c)
		return
	}
	response.OkWithMessage("删除标签成功", c)
}
