package blog

import (
	"github.com/gin-gonic/gin"
	blogApi "github.com/isgvto/gin-vue-admin-gblog/server/api/v1/blog"
)

type AiRouter struct{}

func (r *AiRouter) InitAiRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	blogApi.StartVisualCleanup()
	aiRouter := Router.Group("blog/ai")
	{
		aiRouter.GET("status", aiApi.Status)
		aiRouter.POST("chat", aiApi.Chat)
		aiRouter.POST("summary", aiApi.Summary)
		aiRouter.POST("suggest-tags", aiApi.SuggestTags)
		aiRouter.GET("visual/status", aiApi.VisualStatus)
		aiRouter.POST("visual/plan", aiApi.VisualPlan)
		aiRouter.POST("visual/generate", aiApi.VisualGenerate)
		aiRouter.POST("visual/adopt", aiApi.VisualAdopt)
		aiRouter.POST("visual/discard", aiApi.VisualDiscard)
	}
}
