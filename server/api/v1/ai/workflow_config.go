package ai

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	aiService "github.com/isgvto/gin-vue-admin-gblog/server/service/ai"
)

// GetWorkflowConfig reads workflow model assignment.
// @Tags AI模型配置
// @Summary 读取需求工作流模型分配
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response
// @Router /ai/modelConfig/workflow [get]
func (a *ModelConfigApi) GetWorkflowConfig(c *gin.Context) {
	cfg, err := (aiService.WorkflowService{}).Config()
	if err != nil {
		response.FailWithMessage("读取需求工作流配置失败", c)
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

// SaveWorkflowConfig saves workflow model assignment.
// @Tags AI模型配置
// @Summary 保存需求工作流模型分配
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body ai.WorkflowConfig true "功能模型配置"
// @Success 200 {object} response.Response
// @Router /ai/modelConfig/workflow [put]
func (a *ModelConfigApi) SaveWorkflowConfig(c *gin.Context) {
	var cfg aiModel.WorkflowConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.FailWithMessage("配置格式错误", c)
		return
	}
	if err := (aiService.WorkflowService{}).Save(cfg); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("需求工作流模型分配已保存", c)
}

// TestWorkflowConfig tests structured generation with the submitted assignment.
// @Tags AI模型配置
// @Summary 测试需求工作流模型
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body ai.WorkflowConfig true "待测试的功能模型配置"
// @Success 200 {object} response.Response
// @Router /ai/modelConfig/workflow/test [post]
func (a *ModelConfigApi) TestWorkflowConfig(c *gin.Context) {
	var cfg aiModel.WorkflowConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.FailWithMessage("配置格式错误", c)
		return
	}
	text, model, err := (aiService.WorkflowService{}).Test(c.Request.Context(), cfg)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"solution": text, "model": model}, c)
}
