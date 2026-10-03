package blog

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"go.uber.org/zap"
)

type SiteApi struct{}

func (a *SiteApi) GetSite(c *gin.Context) {
	data, err := siteService.GetSiteInfo()
	if err != nil {
		global.GVA_LOG.Error("get site info failed", zap.Error(err))
		response.FailWithMessage("获取站点信息失败", c)
		return
	}
	response.OkWithData(data, c)
}
