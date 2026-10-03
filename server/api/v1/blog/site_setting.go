package blog

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"go.uber.org/zap"
)

type SiteSettingApi struct{}

func (a *SiteSettingApi) GetSiteSettings(c *gin.Context) {
	data, err := siteSettingService.GetGrouped()
	if err != nil {
		global.GVA_LOG.Error("get site settings failed", zap.Error(err))
		response.FailWithMessage("获取站点配置失败", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *SiteSettingApi) UpdateSiteSettings(c *gin.Context) {
	var req blogReq.SiteSettingBatchUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := siteSettingService.UpdateAll(req); err != nil {
		global.GVA_LOG.Error("update site settings failed", zap.Error(err))
		response.FailWithMessage("更新站点配置失败", c)
		return
	}
	response.OkWithMessage("更新站点配置成功", c)
}

func (a *SiteSettingApi) GetWebTitleSuffix(c *gin.Context) {
	value, err := siteSettingService.GetWebTitleSuffix()
	if err != nil {
		global.GVA_LOG.Error("get web title suffix failed", zap.Error(err))
		response.FailWithMessage("获取网页标题后缀失败", c)
		return
	}
	response.OkWithData(value, c)
}
