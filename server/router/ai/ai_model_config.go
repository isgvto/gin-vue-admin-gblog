package ai

import "github.com/gin-gonic/gin"

type ModelConfigRouter struct{}

func (r *ModelConfigRouter) InitModelConfigRouter(Router *gin.RouterGroup) {
	configRouter := Router.Group("ai/modelConfig")
	{
		configRouter.GET("list", modelConfigApi.GetList)
		configRouter.GET("providers", modelConfigApi.Providers)
		configRouter.GET("errorAnalysis", modelConfigApi.GetErrorAnalysis)
		configRouter.GET("workflow", modelConfigApi.GetWorkflowConfig)
		configRouter.PUT("workflow", modelConfigApi.SaveWorkflowConfig)
		configRouter.POST("workflow/test", modelConfigApi.TestWorkflowConfig)
		configRouter.PUT("errorAnalysis", modelConfigApi.SaveErrorAnalysis)
		configRouter.POST("errorAnalysis/test", modelConfigApi.TestErrorAnalysis)
		configRouter.POST("testConnection", modelConfigApi.TestConnection)
		configRouter.POST("providerModels", modelConfigApi.ProviderModels)
		configRouter.POST("", modelConfigApi.Create)
		configRouter.PUT("", modelConfigApi.Update)
		configRouter.DELETE(":id", modelConfigApi.Delete)
		configRouter.PUT("setDefault/:id", modelConfigApi.SetDefault)
	}
}
