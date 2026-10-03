package blog

import (
	"github.com/gin-gonic/gin"
	blogmw "github.com/isgvto/gin-vue-admin-gblog/server/middleware/blog"
)

type FriendRouter struct{}

func (r *FriendRouter) InitFriendRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	adminRecordRouter := Router.Group("admin").Use(blogmw.OperationRecord())
	adminRouter := Router.Group("admin")
	PublicRouter.Group("").Use(blogmw.VisitRecord(blogmw.VisitBehaviorFriend)).GET("friends", friendApi.GetFriends)
	PublicRouter.Group("").Use(blogmw.VisitRecord(blogmw.VisitBehaviorClickFriend)).POST("friend", friendApi.AddFriendViews)
	adminRouter.GET("friends", friendApi.GetFriendList)
	adminRouter.GET("friendInfo", friendApi.GetFriendInfo)
	adminRecordRouter.PUT("friend/published", friendApi.UpdateFriendPublished)
	adminRecordRouter.POST("friend", friendApi.CreateFriend)
	adminRecordRouter.PUT("friend", friendApi.UpdateFriend)
	adminRecordRouter.DELETE("friend", friendApi.DeleteFriend)
	adminRecordRouter.PUT("friendInfo/commentEnabled", friendApi.UpdateFriendInfoCommentEnabled)
	adminRecordRouter.PUT("friendInfo/content", friendApi.UpdateFriendInfoContent)
}
