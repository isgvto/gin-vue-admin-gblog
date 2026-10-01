package blog

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"go.uber.org/zap"
)

type ArchiveApi struct{}

func (a *ArchiveApi) GetArchives(c *gin.Context) {
	var req blogReq.ArchiveSearch
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	data, err := archiveService.GetArchive(req)
	if err != nil {
		global.GVA_LOG.Error("get archives failed", zap.Error(err))
		response.FailWithMessage("获取归档失败", c)
		return
	}
	response.OkWithData(data, c)
}
