package ai

import api "github.com/isgvto/gin-vue-admin-gblog/server/api/v1"

type RouterGroup struct {
	ModelConfigRouter
}

var modelConfigApi = api.ApiGroupApp.AiApiGroup.ModelConfigApi
