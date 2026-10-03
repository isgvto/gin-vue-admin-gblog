package blog

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogmw "github.com/isgvto/gin-vue-admin-gblog/server/middleware/blog"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	blogService "github.com/isgvto/gin-vue-admin-gblog/server/service/blog"
	"go.uber.org/zap"
)

type ArticleApi struct{}

func (a *ArticleApi) GetArticleList(c *gin.Context) {
	var req blogReq.ArticleSearch
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := articleService.GetPublishedListView(req)
	if err != nil {
		global.GVA_LOG.Error("get article list failed", zap.Error(err))
		response.FailWithMessage("获取文章列表失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "获取文章列表成功", c)
}

func (a *ArticleApi) GetArticle(c *gin.Context) {
	var req blogReq.IDQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	data, err := articleService.GetPublishedDetailByIDWithToken(req.ID, c.GetHeader("Authorization"))
	if err != nil {
		if errors.Is(err, blogService.ErrPasswordProtected) {
			response.FailWithMessage("文章受密码保护", c)
			return
		}
		global.GVA_LOG.Error("get article failed", zap.Error(err))
		response.FailWithMessage("获取文章失败", c)
		return
	}
	blogmw.SetVisitContent(c, data.Title)
	response.OkWithData(data, c)
}

func (a *ArticleApi) SearchBlog(c *gin.Context) {
	var req blogReq.SearchBlogQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	data, err := articleService.SearchPublishedBlogs(req.Query)
	if err != nil {
		global.GVA_LOG.Error("search blog failed", zap.Error(err))
		response.FailWithMessage("获取搜索结果失败", c)
		return
	}
	response.OkWithData(data, c)
}
