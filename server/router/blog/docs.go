package blog

import (
	"github.com/gin-gonic/gin"
	blogmw "github.com/isgvto/gin-vue-admin-gblog/server/middleware/blog"
)

type DocsRouter struct{}

func (r *DocsRouter) InitDocsRouter(PrivateRouter *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	PublicRouter.GET("docs/tree", docsApi.GetTree)
	PublicRouter.GET("docs/content", docsApi.GetContent)
	PublicRouter.POST("docs/webhook", docsApi.Webhook)
	PrivateRouter.Group("admin").Use(blogmw.OperationRecord()).POST("docs/sync", docsApi.Sync)
}
