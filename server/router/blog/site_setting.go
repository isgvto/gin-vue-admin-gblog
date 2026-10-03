package blog

import (
	"github.com/gin-gonic/gin"
	blogmw "github.com/isgvto/gin-vue-admin-gblog/server/middleware/blog"
)

type SiteSettingRouter struct{}

func (r *SiteSettingRouter) InitSiteSettingRouter(Router *gin.RouterGroup) {
	adminRecordRouter := Router.Group("admin").Use(blogmw.OperationRecord())
	adminRouter := Router.Group("admin")
	adminRouter.GET("siteSettings", siteSettingApi.GetSiteSettings)
	adminRouter.GET("webTitleSuffix", siteSettingApi.GetWebTitleSuffix)
	adminRecordRouter.POST("siteSettings", siteSettingApi.UpdateSiteSettings)
}
