package blog

import (
	"github.com/gin-gonic/gin"
	blogmw "github.com/isgvto/gin-vue-admin-gblog/server/middleware/blog"
)

type CommentRouter struct{}

func (r *CommentRouter) InitCommentRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	adminRecordRouter := Router.Group("admin").Use(blogmw.OperationRecord())
	adminRouter := Router.Group("admin")
	PublicRouter.GET("comments", commentApi.GetComments)
	PublicRouter.POST("comment", commentApi.CreateComment)
	adminRouter.GET("comments", commentApi.GetCommentList)
	adminRouter.GET("blogIdAndTitle", commentApi.GetBlogIDAndTitle)
	adminRecordRouter.PUT("comment/published", commentApi.UpdateCommentPublished)
	adminRecordRouter.PUT("comment/notice", commentApi.UpdateCommentNotice)
	adminRecordRouter.PUT("comment", commentApi.UpdateComment)
	adminRecordRouter.DELETE("comment", commentApi.DeleteComment)
}
