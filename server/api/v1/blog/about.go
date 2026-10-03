package blog

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"go.uber.org/zap"
)

type AboutApi struct{}

func (a *AboutApi) GetAbout(c *gin.Context) {
	var (
		list any
		err  error
	)
	if strings.Contains(c.FullPath(), "admin/") {
		list, err = aboutService.GetList()
	} else {
		list, err = aboutService.GetPublicList()
	}
	if err != nil {
		global.GVA_LOG.Error("get about failed", zap.Error(err))
		response.FailWithMessage("获取关于页失败", c)
		return
	}
	response.OkWithData(list, c)
}

func (a *AboutApi) UpdateAbout(c *gin.Context) {
	var req blogReq.AboutUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := aboutService.UpdateValues(req.Values); err != nil {
		global.GVA_LOG.Error("update about failed", zap.Error(err))
		response.FailWithMessage("更新关于页失败", c)
		return
	}
	response.OkWithMessage("更新关于页成功", c)
}
