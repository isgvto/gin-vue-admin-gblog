package service

import (
	"github.com/isgvto/gin-vue-admin-gblog/server/service/ai"
	"github.com/isgvto/gin-vue-admin-gblog/server/service/blog"
	"github.com/isgvto/gin-vue-admin-gblog/server/service/example"
	"github.com/isgvto/gin-vue-admin-gblog/server/service/system"
)

var ServiceGroupApp = new(ServiceGroup)

type ServiceGroup struct {
	SystemServiceGroup  system.ServiceGroup
	ExampleServiceGroup example.ServiceGroup
	BlogServiceGroup    blog.ServiceGroup
	AiServiceGroup      ai.ServiceGroup
}
