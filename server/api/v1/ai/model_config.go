package ai

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	aiReq "github.com/isgvto/gin-vue-admin-gblog/server/model/ai/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"github.com/isgvto/gin-vue-admin-gblog/server/service"
	"go.uber.org/zap"
)

type ModelConfigApi struct{}

var modelConfigService = service.ServiceGroupApp.AiServiceGroup.ModelConfigService

func (a *ModelConfigApi) GetList(c *gin.Context) {
	var req aiReq.AiModelConfigSearch
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := modelConfigService.GetList(req)
	if err != nil {
		global.GVA_LOG.Error("获取 AI 模型配置失败", zap.Error(err))
		response.FailWithMessage("获取 AI 模型配置失败", c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total, "page": req.Page, "pageSize": req.PageSize}, "获取成功", c)
}

func (a *ModelConfigApi) Create(c *gin.Context) {
	var req aiReq.AiModelConfigUpsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := modelConfigService.Create(req); err != nil {
		global.GVA_LOG.Error("创建 AI 模型配置失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

func (a *ModelConfigApi) Update(c *gin.Context) {
	var req aiReq.AiModelConfigUpsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := modelConfigService.Update(req); err != nil {
		global.GVA_LOG.Error("更新 AI 模型配置失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (a *ModelConfigApi) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("无效的配置 ID", c)
		return
	}
	if err := modelConfigService.Delete(uint(id)); err != nil {
		global.GVA_LOG.Error("删除 AI 模型配置失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *ModelConfigApi) SetDefault(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("无效的配置 ID", c)
		return
	}
	if err := modelConfigService.SetDefault(uint(id)); err != nil {
		global.GVA_LOG.Error("设置默认模型失败", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("设置成功", c)
}

func (a *ModelConfigApi) Providers(c *gin.Context) {
	response.OkWithData(modelConfigService.Providers(), c)
}

func (a *ModelConfigApi) TestConnection(c *gin.Context) {
	var req aiReq.AiModelConfigUpsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := modelConfigService.TestConnection(c.Request.Context(), req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("连接成功，模型已返回有效响应", c)
}

func (a *ModelConfigApi) ProviderModels(c *gin.Context) {
	var req aiReq.AiModelConfigUpsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	models, err := modelConfigService.ListProviderModels(c.Request.Context(), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(models, c)
}
