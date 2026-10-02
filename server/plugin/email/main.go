package email

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/plugin/email/config"
	"github.com/isgvto/gin-vue-admin-gblog/server/plugin/email/global"
	"github.com/isgvto/gin-vue-admin-gblog/server/plugin/email/router"
)

type emailPlugin struct{}

func CreateEmailPlug(To, From, Host, Secret, Nickname string, Port int, IsSSL bool, IsLoginAuth bool) *emailPlugin {
	global.ConfigProvider = nil
	global.GlobalConfig.To = To
	global.GlobalConfig.From = From
	global.GlobalConfig.Host = Host
	global.GlobalConfig.Secret = Secret
	global.GlobalConfig.Nickname = Nickname
	global.GlobalConfig.Port = Port
	global.GlobalConfig.IsSSL = IsSSL
	global.GlobalConfig.IsLoginAuth = IsLoginAuth
	return &emailPlugin{}
}

// CreateEmailPlugWithConfigProvider 在每次发送时获取当前配置，避免热更新后使用启动时的副本。
func CreateEmailPlugWithConfigProvider(provider func() config.Email) *emailPlugin {
	global.ConfigProvider = provider
	return &emailPlugin{}
}

func (*emailPlugin) Register(group *gin.RouterGroup) {
	router.RouterGroupApp.InitEmailRouter(group)
}

func (*emailPlugin) RouterPath() string {
	return "email"
}
