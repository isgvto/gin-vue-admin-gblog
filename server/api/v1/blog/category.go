package blog

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"go.uber.org/zap"
)

type CategoryApi struct{}

func (a *CategoryApi) GetCategoryList(c *gin.Context) {
	list, err := categoryService.GetList()
	if err != nil {
		global.GVA_LOG.Error("get category list failed", zap.Error(err))
		response.FailWithMessage("获取分类列表失败", c)
		return
	}
	response.OkWithData(list, c)
}

func (a *CategoryApi) GetCategoryBlogList(c *gin.Context) {
	var req blogReq.CategoryBlogSearch
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := categoryService.GetPublishedBlogListByName(req)
	if err != nil {
		global.GVA_LOG.Error("get category blog list failed", zap.Error(err))
		response.FailWithMessage("鑾峰彇鍒嗙被鏂囩珷鍒楄〃澶辫触", c)
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
	}, "鑾峰彇鍒嗙被鏂囩珷鍒楄〃鎴愬姛", c)
}

func (a *CategoryApi) CreateCategory(c *gin.Context) {
	var req blogReq.CategoryUpsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := categoryService.Create(req); err != nil {
		global.GVA_LOG.Error("create category failed", zap.Error(err))
		response.FailWithMessage("创建分类失败", c)
		return
	}
	response.OkWithMessage("创建分类成功", c)
}

func (a *CategoryApi) UpdateCategory(c *gin.Context) {
	var req blogReq.CategoryUpsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := categoryService.Update(req); err != nil {
		global.GVA_LOG.Error("update category failed", zap.Error(err))
		response.FailWithMessage("更新分类失败", c)
		return
	}
	response.OkWithMessage("更新分类成功", c)
}

func (a *CategoryApi) DeleteCategory(c *gin.Context) {
	var req blogReq.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := categoryService.Delete(req.ID); err != nil {
		global.GVA_LOG.Error("delete category failed", zap.Error(err))
		response.FailWithMessage("删除分类失败", c)
		return
	}
	response.OkWithMessage("删除分类成功", c)
}
