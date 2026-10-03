package blog

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	blogService "github.com/isgvto/gin-vue-admin-gblog/server/service/blog"
	"github.com/isgvto/gin-vue-admin-gblog/server/utils"
)

var visualCleanupOnce sync.Once

// VisualStatus 配图能力与当前存储状态
// @Tags BlogAI
// @Summary 获取配图能力状态
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /blog/ai/visual/status [get]
func (*AiApi) VisualStatus(c *gin.Context) {
	response.OkWithData((&blogService.VisualService{}).Status(), c)
}

// VisualPlan 根据文章推荐类型和可编辑描述
// @Tags BlogAI
// @Summary 推荐图示或配图方案
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body blogService.VisualRequest true "文章与配图要求"
// @Success 200 {object} response.Response{data=blogService.VisualPlan}
// @Router /blog/ai/visual/plan [post]
func (*AiApi) VisualPlan(c *gin.Context) { visualRequest(c, false) }

// VisualGenerate 生成私有临时预览，不上传对象存储
// @Tags BlogAI
// @Summary 生成图示或配图预览
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body blogService.VisualRequest true "已确认的配图类型和要求"
// @Success 200 {object} response.Response{data=blogService.VisualResult}
// @Router /blog/ai/visual/generate [post]
func (*AiApi) VisualGenerate(c *gin.Context) { visualRequest(c, true) }
func visualRequest(c *gin.Context, generate bool) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		response.FailWithMessage("请先登录", c)
		return
	}
	var req blogService.VisualRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 256<<10)
	if c.ShouldBindJSON(&req) != nil {
		response.FailWithMessage("配图参数无效", c)
		return
	}
	if err := blogService.ValidateVisualRequest(req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if generate && req.Kind == "auto" {
		response.FailWithMessage("请先推荐类型并确认，再生成预览", c)
		return
	}
	if _, err := aiService.ConsumeQuota(c.Request.Context(), userID); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	service := &blogService.VisualService{}
	if generate {
		data, err := service.Generate(c.Request.Context(), userID, req)
		if err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
		response.OkWithData(data, c)
	} else {
		data, err := service.Plan(c.Request.Context(), req)
		if err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
		response.OkWithData(data, c)
	}
}

// VisualAdopt 验证预览归属并上传当前存储，返回附件链接
// @Tags BlogAI
// @Summary 采用配图并上传
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body blogService.VisualAdoptRequest true "预览ID；图示需附导出的PNG"
// @Success 200 {object} response.Response
// @Router /blog/ai/visual/adopt [post]
func (*AiApi) VisualAdopt(c *gin.Context) {
	var req blogService.VisualAdoptRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 15<<20)
	if c.ShouldBindJSON(&req) != nil {
		response.FailWithMessage("上传参数无效", c)
		return
	}
	data, err := (&blogService.VisualService{}).Adopt(c.Request.Context(), utils.GetUserID(c), req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(data, c)
}

// VisualDiscard 删除当前用户未上传的临时预览
// @Tags BlogAI
// @Summary 丢弃配图预览
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param data body blogService.VisualAdoptRequest true "预览ID"
// @Success 200 {object} response.Response
// @Router /blog/ai/visual/discard [post]
func (*AiApi) VisualDiscard(c *gin.Context) {
	var req blogService.VisualAdoptRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024)
	if c.ShouldBindJSON(&req) != nil {
		response.FailWithMessage("预览标识无效", c)
		return
	}
	if err := (&blogService.VisualService{}).Discard(utils.GetUserID(c), req.ID); err != nil {
		response.FailWithMessage("丢弃预览失败", c)
		return
	}
	response.Ok(c)
}
func StartVisualCleanup() {
	visualCleanupOnce.Do(func() {
		go func() {
			for {
				if global.GVA_DB != nil && global.GVA_DB.Migrator().HasTable("blog_ai_visual_tasks") {
					_ = blogService.CleanupVisualTasks(global.GVA_DB)
				}
				time.Sleep(time.Hour)
			}
		}()
	})
}
