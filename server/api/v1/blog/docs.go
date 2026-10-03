package blog

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"go.uber.org/zap"
)

type DocsApi struct{}

func (a *DocsApi) GetTree(c *gin.Context) {
	data, err := docsService.GetTree()
	if err != nil {
		global.GVA_LOG.Error("get docs tree failed", zap.Error(err))
		response.FailWithMessage("get docs tree failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *DocsApi) GetContent(c *gin.Context) {
	var req blogReq.DocContentQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	data, err := docsService.GetContent(req.Path)
	if err != nil {
		global.GVA_LOG.Error("get doc content failed", zap.Error(err))
		response.FailWithMessage("get doc content failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *DocsApi) Sync(c *gin.Context) {
	data, err := docsService.SyncTree()
	if err != nil {
		global.GVA_LOG.Error("sync docs failed", zap.Error(err))
		response.FailWithMessage("sync docs failed: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(data, "sync docs success", c)
}

func (a *DocsApi) Webhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.FailWithMessage("read webhook body failed", c)
		return
	}
	if err := docsService.ValidateWebhookSignature(body, c.GetHeader("X-Hub-Signature-256")); err != nil {
		response.NoAuth(err.Error(), c)
		return
	}
	if _, err := docsService.SyncTree(); err != nil {
		global.GVA_LOG.Error("sync docs from webhook failed", zap.Error(err))
		response.FailWithMessage("sync docs failed: "+err.Error(), c)
		return
	}
	response.OkWithMessage("sync docs success", c)
}
