package global

import "github.com/isgvto/gin-vue-admin-gblog/server/plugin/email/config"

var GlobalConfig = new(config.Email)

// ConfigProvider 在插件初始化时设置，使发送端能够读取主程序的最新配置。
// 未设置时保持独立插件通过 GlobalConfig 配置的兼容行为。
var ConfigProvider func() config.Email

func GetConfig() config.Email {
	if ConfigProvider != nil {
		return ConfigProvider()
	}
	return *GlobalConfig
}
