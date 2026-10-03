package ai

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	aiService "github.com/isgvto/gin-vue-admin-gblog/server/service/ai"
)

// GetErrorAnalysis reads the feature binding and safe model choices.
// @Tags AI模型配置
// @Summary 读取错误分析模型分配
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response
// @Router /ai/modelConfig/errorAnalysis [get]
func (a *ModelConfigApi) GetErrorAnalysis(c *gin.Context) {
	cfg, err := (aiService.ErrorAnalysisService{}).Config()
	if err != nil {
		response.FailWithMessage("读取错误分析配置失败", c)
		return
	}
	var models []struct {
		ID    uint   `json:"id"`
		Name  string `json:"name"`
		Model string `json:"model"`
	}
	if err := global.GVA_DB.Model(&aiModel.AiModelConfig{}).Select("id", "name", "model").Where("status = ?", true).Order("id ASC").Find(&models).Error; err != nil {
		response.FailWithMessage("读取可用模型失败", c)
		return
	}
	response.OkWithData(gin.H{"config": cfg, "models": models}, c)
}

// SaveErrorAnalysis saves the feature binding.
// @Tags AI模型配置
// @Summary 保存错误分析模型分配
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body ai.ErrorAnalysisConfig true "错误分析配置"
// @Success 200 {object} response.Response
// @Router /ai/modelConfig/errorAnalysis [put]
func (a *ModelConfigApi) SaveErrorAnalysis(c *gin.Context) {
	var cfg aiModel.ErrorAnalysisConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.FailWithMessage("配置格式错误", c)
		return
	}
	if err := (aiService.ErrorAnalysisService{}).Save(cfg); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("错误分析配置已保存", c)
}

// TestErrorAnalysis executes the same analyzer with a fixed sample.
// @Tags AI模型配置
// @Summary 使用当前分配进行示例错误分析
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body ai.ErrorAnalysisConfig true "待测试的错误分析配置"
// @Success 200 {object} response.Response
// @Router /ai/modelConfig/errorAnalysis/test [post]
func (a *ModelConfigApi) TestErrorAnalysis(c *gin.Context) {
	var cfg aiModel.ErrorAnalysisConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.FailWithMessage("配置格式错误", c)
		return
	}
	svc := aiService.ErrorAnalysisService{}
	model, err := svc.Resolve(cfg)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	result, err := svc.Analyze(c.Request.Context(), cfg, model, "Go / GORM 示例", "查询用户列表失败：Error 1146 (42S02): Table 'example.users' doesn't exist")
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"solution": result, "model": model.Name + " / " + model.Model}, c)
}
