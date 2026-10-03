package blog

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"go.uber.org/zap"
)

type DashboardApi struct{}

func (a *DashboardApi) GetDashboard(c *gin.Context) {
	data, err := dashboardService.GetSummary()
	if err != nil {
		global.GVA_LOG.Error("get dashboard failed", zap.Error(err))
		response.FailWithMessage("获取仪表盘数据失败", c)
		return
	}
	response.OkWithData(data, c)
}
