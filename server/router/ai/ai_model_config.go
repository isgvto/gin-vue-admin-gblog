package ai

import "github.com/gin-gonic/gin"

type ModelConfigRouter struct{}

func (r *ModelConfigRouter) InitModelConfigRouter(Router *gin.RouterGroup) {
	configRouter := Router.Group("ai/modelConfig")
	{
		configRouter.GET("list", modelConfigApi.GetList)
		configRouter.GET("providers", modelConfigApi.Providers)
		configRouter.POST("testConnection", modelConfigApi.TestConnection)
		configRouter.POST("providerModels", modelConfigApi.ProviderModels)
		configRouter.POST("", modelConfigApi.Create)
		configRouter.PUT("", modelConfigApi.Update)
		configRouter.DELETE(":id", modelConfigApi.Delete)
		configRouter.PUT("setDefault/:id", modelConfigApi.SetDefault)
	}
}
