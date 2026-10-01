package blog

import (
	"github.com/gin-gonic/gin"
	blogmw "github.com/isgvto/gin-vue-admin-gblog/server/middleware/blog"
)

type AboutRouter struct{}

func (r *AboutRouter) InitAboutRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	adminRecordRouter := Router.Group("admin").Use(blogmw.OperationRecord())
	adminRouter := Router.Group("admin")
	PublicRouter.Group("").Use(blogmw.VisitRecord(blogmw.VisitBehaviorAbout)).GET("about", aboutApi.GetAbout)
	adminRouter.GET("about", aboutApi.GetAbout)
	adminRecordRouter.PUT("about", aboutApi.UpdateAbout)
}
