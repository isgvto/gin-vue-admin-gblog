package initialize

import (
	_ "github.com/isgvto/gin-vue-admin-gblog/server/source/blog"
	_ "github.com/isgvto/gin-vue-admin-gblog/server/source/example"
	_ "github.com/isgvto/gin-vue-admin-gblog/server/source/system"
)

func init() {
	// do nothing,only import source package so that inits can be registered
}
