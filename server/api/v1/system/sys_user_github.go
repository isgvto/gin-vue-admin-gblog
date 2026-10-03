package system

import (
	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	systemService "github.com/isgvto/gin-vue-admin-gblog/server/service/system"
	"github.com/isgvto/gin-vue-admin-gblog/server/utils"
	"strings"
)

func (b *BaseApi) SetGitHubProfile(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
	}
	if c.ShouldBindJSON(&input) != nil {
		response.FailWithMessage("请输入 GitHub 用户名", c)
		return
	}
	username := strings.TrimSpace(input.Username)
	if username != "" && !systemService.ValidGitHubUsername(username) {
		response.FailWithMessage("GitHub 用户名格式无效", c)
		return
	}
	result := global.GVA_DB.Model(&system.SysUser{}).Where("id = ?", utils.GetUserID(c)).Update("github_username", username)
	if result.Error != nil {
		response.FailWithMessage("保存 GitHub 展示账号失败", c)
		return
	}
	response.OkWithDetailed(gin.H{"username": username}, "已保存展示账号", c)
}

func (b *BaseApi) GetGitHubProfile(c *gin.Context) {
	var user system.SysUser
	if err := global.GVA_DB.Select("id", "github_username").First(&user, utils.GetUserID(c)).Error; err != nil {
		response.FailWithMessage("无法读取个人 GitHub 配置", c)
		return
	}
	if user.GitHubUsername == "" {
		response.OkWithData(gin.H{"username": ""}, c)
		return
	}
	data, err := systemService.ReadGitHubProfile(c.Request.Context(), user.GitHubUsername)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(data, c)
}
