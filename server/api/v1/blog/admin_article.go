package blog

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"go.uber.org/zap"
)

type AdminArticleApi struct{}

func (a *AdminArticleApi) GetArticleList(c *gin.Context) {
	var req blogReq.AdminArticleSearch
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := adminArticleService.GetListView(req)
	if err != nil {
		global.GVA_LOG.Error("get admin article list failed", zap.Error(err))
		response.FailWithMessage("获取后台文章列表失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "获取后台文章列表成功", c)
}

func (a *AdminArticleApi) GetArticle(c *gin.Context) {
	var req blogReq.IDQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	data, err := adminArticleService.GetDetailByID(req.ID)
	if err != nil {
		global.GVA_LOG.Error("get admin article failed", zap.Error(err))
		response.FailWithMessage("获取后台文章失败", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *AdminArticleApi) GetCategoryAndTag(c *gin.Context) {
	data, err := adminArticleService.GetCategoryAndTag()
	if err != nil {
		global.GVA_LOG.Error("get category and tag failed", zap.Error(err))
		response.FailWithMessage("获取分类和标签失败", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *AdminArticleApi) CreateArticle(c *gin.Context) {
	var req blogReq.ArticleUpsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := adminArticleService.Create(req); err != nil {
		global.GVA_LOG.Error("create article failed", zap.Error(err))
		response.FailWithMessage("创建文章失败", c)
		return
	}
	response.OkWithMessage("创建文章成功", c)
}

func (a *AdminArticleApi) UpdateArticle(c *gin.Context) {
	var req blogReq.ArticleUpsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := adminArticleService.Update(req); err != nil {
		global.GVA_LOG.Error("update article failed", zap.Error(err))
		response.FailWithMessage("更新文章失败", c)
		return
	}
	response.OkWithMessage("更新文章成功", c)
}

func (a *AdminArticleApi) DeleteArticle(c *gin.Context) {
	var req blogReq.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := adminArticleService.Delete(req.ID); err != nil {
		global.GVA_LOG.Error("delete article failed", zap.Error(err))
		response.FailWithMessage("删除文章失败", c)
		return
	}
	response.OkWithMessage("删除文章成功", c)
}

func (a *AdminArticleApi) UpdateArticleTop(c *gin.Context) {
	var req blogReq.ToggleReq
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := adminArticleService.UpdateTop(req.ID, req.Value); err != nil {
		global.GVA_LOG.Error("update article top failed", zap.Error(err))
		response.FailWithMessage("更新文章置顶状态失败", c)
		return
	}
	response.OkWithMessage("更新文章置顶状态成功", c)
}

func (a *AdminArticleApi) UpdateArticleRecommend(c *gin.Context) {
	var req blogReq.ToggleReq
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := adminArticleService.UpdateRecommend(req.ID, req.Value); err != nil {
		global.GVA_LOG.Error("update article recommend failed", zap.Error(err))
		response.FailWithMessage("更新文章推荐状态失败", c)
		return
	}
	response.OkWithMessage("更新文章推荐状态成功", c)
}

func (a *AdminArticleApi) UpdateArticleVisibility(c *gin.Context) {
	var uriReq blogReq.IDUriReq
	if err := c.ShouldBindUri(&uriReq); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var req blogReq.BlogVisibility
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := adminArticleService.UpdateVisibility(uriReq.ID, req); err != nil {
		global.GVA_LOG.Error("update article visibility failed", zap.Error(err))
		response.FailWithMessage("更新文章可见性失败", c)
		return
	}
	response.OkWithMessage("更新文章可见性成功", c)
}
