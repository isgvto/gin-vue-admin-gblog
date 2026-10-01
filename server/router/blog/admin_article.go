package blog

import (
	"github.com/gin-gonic/gin"
	blogmw "github.com/isgvto/gin-vue-admin-gblog/server/middleware/blog"
)

type AdminArticleRouter struct{}

func (r *AdminArticleRouter) InitAdminArticleRouter(Router *gin.RouterGroup) {
	adminRecordRouter := Router.Group("admin").Use(blogmw.OperationRecord())
	adminRouter := Router.Group("admin")
	{
		adminRouter.GET("blogs", adminArticleApi.GetArticleList)
		adminRouter.GET("blog", adminArticleApi.GetArticle)
		adminRouter.GET("categoryAndTag", adminArticleApi.GetCategoryAndTag)
	}
	{
		adminRecordRouter.POST("blog", adminArticleApi.CreateArticle)
		adminRecordRouter.PUT("blog", adminArticleApi.UpdateArticle)
		adminRecordRouter.DELETE("blog", adminArticleApi.DeleteArticle)
		adminRecordRouter.PUT("blog/top", adminArticleApi.UpdateArticleTop)
		adminRecordRouter.PUT("blog/recommend", adminArticleApi.UpdateArticleRecommend)
		adminRecordRouter.PUT("blog/:id/visibility", adminArticleApi.UpdateArticleVisibility)
	}
}
