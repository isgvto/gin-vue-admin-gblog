package blog

import (
	"github.com/gin-gonic/gin"
	blogmw "github.com/isgvto/gin-vue-admin-gblog/server/middleware/blog"
)

type ArchiveRouter struct{}

func (r *ArchiveRouter) InitArchiveRouter(PublicRouter *gin.RouterGroup) {
	PublicRouter.Group("").Use(blogmw.VisitRecord(blogmw.VisitBehaviorArchive)).GET("archives", archiveApi.GetArchives)
}
