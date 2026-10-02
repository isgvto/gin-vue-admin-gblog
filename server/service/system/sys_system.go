package system

import (
	"github.com/isgvto/gin-vue-admin-gblog/server/config"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	"github.com/isgvto/gin-vue-admin-gblog/server/utils"
	"go.uber.org/zap"
)

//@author: [piexlmax](https://github.com/piexlmax)
//@function: GetSystemConfig
//@description: 读取配置文件
//@return: conf config.Server, err error

type SystemConfigService struct{}

var SystemConfigServiceApp = new(SystemConfigService)

func (systemConfigService *SystemConfigService) GetSystemConfig() (conf config.Server, err error) {
	return global.GVA_CONFIG, nil
}

// @description   set system config,
//@author: [piexlmax](https://github.com/piexlmax)
//@function: SetSystemConfig
//@description: 设置配置文件
//@param: system model.System
//@return: err error

func (systemConfigService *SystemConfigService) SetSystemConfig(system system.System) (err error) {
	cs := utils.StructToMap(system.Config)
	for k, v := range cs {
		global.GVA_VP.Set(k, v)
	}
	err = global.GVA_VP.WriteConfig()
	if err == nil {
		// 邮件测试可能紧接着保存请求到达，不等待文件监听器异步刷新。
		global.GVA_CONFIG.Email = system.Config.Email
	}
	return err
}

//@author: [SliverHorn](https://github.com/SliverHorn)
//@function: GetServerInfo
//@description: 获取服务器信息
//@return: server *utils.Server, err error

func (systemConfigService *SystemConfigService) GetServerInfo() (server *utils.Server, err error) {
	var s utils.Server
	s.Os = utils.InitOS()
	if s.Cpu, err = utils.InitCPU(); err != nil {
		global.GVA_LOG.Error("func utils.InitCPU() Failed", zap.String("err", err.Error()))
		return &s, err
	}
	if s.Ram, err = utils.InitRAM(); err != nil {
		global.GVA_LOG.Error("func utils.InitRAM() Failed", zap.String("err", err.Error()))
		return &s, err
	}
	if s.Disk, err = utils.InitDisk(); err != nil {
		global.GVA_LOG.Error("func utils.InitDisk() Failed", zap.String("err", err.Error()))
		return &s, err
	}

	return &s, nil
}
