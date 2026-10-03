package blog

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"go.uber.org/zap"
)

type AuthApi struct{}

func (a *AuthApi) CheckBlogPassword(c *gin.Context) {
	var req blogReq.BlogPasswordCheck
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	token, err := authService.CreateBlogAccessToken(req.BlogID, req.Password)
	if err != nil {
		global.GVA_LOG.Error("check blog password failed", zap.Error(err))
		response.FailWithMessage("密码错误", c)
		return
	}
	response.OkWithData(token, c)
}
