package blog

import (
	"github.com/gin-gonic/gin"
	blogmw "github.com/isgvto/gin-vue-admin-gblog/server/middleware/blog"
)

type AuthRouter struct{}

func (r *AuthRouter) InitAuthRouter(PublicRouter *gin.RouterGroup) {
	PublicRouter.Group("").Use(blogmw.VisitRecord(blogmw.VisitBehaviorCheckPassword)).POST("checkBlogPassword", authApi.CheckBlogPassword)
}
