package v1

import (
	"github.com/isgvto/gin-vue-admin-gblog/server/api/v1/ai"
	"github.com/isgvto/gin-vue-admin-gblog/server/api/v1/blog"
	"github.com/isgvto/gin-vue-admin-gblog/server/api/v1/example"
	"github.com/isgvto/gin-vue-admin-gblog/server/api/v1/system"
)

var ApiGroupApp = new(ApiGroup)

type ApiGroup struct {
	SystemApiGroup  system.ApiGroup
	ExampleApiGroup example.ApiGroup
	BlogApiGroup    blog.ApiGroup
	AiApiGroup      ai.ApiGroup
}
