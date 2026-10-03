package router

import (
	"github.com/isgvto/gin-vue-admin-gblog/server/router/ai"
	"github.com/isgvto/gin-vue-admin-gblog/server/router/blog"
	"github.com/isgvto/gin-vue-admin-gblog/server/router/example"
	"github.com/isgvto/gin-vue-admin-gblog/server/router/system"
)

var RouterGroupApp = new(RouterGroup)

type RouterGroup struct {
	System  system.RouterGroup
	Example example.RouterGroup
	Blog    blog.RouterGroup
	Ai      ai.RouterGroup
}
