package ai

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	aiService "github.com/isgvto/gin-vue-admin-gblog/server/service/ai"
	"net/http"
)

// GetImageConfig 读取配置，不返回密钥
// @Tags AiModelConfig
// @Summary 获取图片生成模型配置
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /ai/modelConfig/image [get]
func (*ModelConfigApi) GetImageConfig(c *gin.Context) {
	cfg, err := (&aiService.ImageService{}).Config()
	if err != nil {
		response.FailWithMessage("读取图片配置失败", c)
		return
	}
	response.OkWithData(gin.H{"config": cfg, "hasKey": cfg.APIKey != ""}, c)
}

// TestImageConfig tests the independent draft without saving or uploading.
// @Tags AiModelConfig
// @Summary 测试独立图片模型配置
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body aiService.ImageConfigInput true "图片模型配置"
// @Success 200 {object} response.Response
// @Router /ai/modelConfig/image/test [post]
func (*ModelConfigApi) TestImageConfig(c *gin.Context) {
	var input aiService.ImageConfigInput
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8192)
	if c.ShouldBindJSON(&input) != nil {
		response.FailWithMessage("图片配置参数无效", c)
		return
	}
	service := &aiService.ImageService{}
	saved, err := service.Config()
	if err != nil {
		response.FailWithMessage("读取图片配置失败", c)
		return
	}
	key := input.APIKey
	if key == "" && !input.ClearKey {
		key = saved.APIKey
	}
	cfg := aiService.ImageEndpointConfig{Provider: input.Provider, BaseURL: input.BaseURL, Model: input.Model, APIKey: key, TimeoutSeconds: input.TimeoutSeconds}
	if cfg.Provider != "ark" && cfg.Provider != "openai" {
		response.FailWithMessage("请选择图片接口协议", c)
		return
	}
	if cfg.Provider == "ark" && cfg.BaseURL == "" {
		cfg.BaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}
	if _, err = service.GenerateWithConfig(c.Request.Context(), cfg, "简洁蓝色几何插画，无文字，用于测试图片接口。", "1024x1024"); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("图片测试成功，未保存配置或上传图片", c)
}

// SaveImageConfig 保存独立图片配置，Key留空时保留已有凭据
// @Tags AiModelConfig
// @Summary 保存图片生成模型配置
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body aiService.ImageConfigInput true "图片模型配置"
// @Success 200 {object} response.Response
// @Router /ai/modelConfig/image [put]
func (*ModelConfigApi) SaveImageConfig(c *gin.Context) {
	var req aiService.ImageConfigInput
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8192)
	if c.ShouldBindJSON(&req) != nil {
		response.FailWithMessage("图片配置参数无效", c)
		return
	}
	if err := (&aiService.ImageService{}).Save(req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("图片配置已保存", c)
}
