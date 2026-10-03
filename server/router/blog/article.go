package blog

import (
	"github.com/gin-gonic/gin"
	blogmw "github.com/isgvto/gin-vue-admin-gblog/server/middleware/blog"
)

type ArticleRouter struct{}

func (r *ArticleRouter) InitArticleRouter(PublicRouter *gin.RouterGroup) {
	PublicRouter.Group("").Use(blogmw.VisitRecord(blogmw.VisitBehaviorIndex)).GET("blogs", articleApi.GetArticleList)
	PublicRouter.Group("").Use(blogmw.VisitRecord(blogmw.VisitBehaviorBlog)).GET("blog", articleApi.GetArticle)
	PublicRouter.Group("").Use(blogmw.VisitRecord(blogmw.VisitBehaviorSearch)).GET("searchBlog", articleApi.SearchBlog)
}
