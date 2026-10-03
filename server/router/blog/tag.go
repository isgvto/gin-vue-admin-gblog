package blog

import (
	"github.com/gin-gonic/gin"
	blogmw "github.com/isgvto/gin-vue-admin-gblog/server/middleware/blog"
)

type TagRouter struct{}

func (r *TagRouter) InitTagRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	adminRecordRouter := Router.Group("admin").Use(blogmw.OperationRecord())
	adminRouter := Router.Group("admin")
	{
		PublicRouter.GET("tag", tagApi.GetTagList)
		PublicRouter.Group("").Use(blogmw.VisitRecord(blogmw.VisitBehaviorTag)).GET("tag/blogs", tagApi.GetTagBlogList)
	}
	{
		adminRouter.GET("tags", tagApi.GetTagList)
	}
	{
		adminRecordRouter.POST("tag", tagApi.CreateTag)
		adminRecordRouter.PUT("tag", tagApi.UpdateTag)
		adminRecordRouter.DELETE("tag", tagApi.DeleteTag)
	}
}
