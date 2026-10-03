package blog

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
)

type TelegramApi struct{}

func (a *TelegramApi) Webhook(c *gin.Context) {
	token := c.Param("token")
	payload := common.JSONMap{}
	_ = c.ShouldBindJSON(&payload)
	response.OkWithData(telegramService.HandleWebhook(token, payload), c)
}
