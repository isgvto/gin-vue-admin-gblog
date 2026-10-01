package router

import "github.com/isgvto/gin-vue-admin-gblog/server/plugin/announcement/api"

var (
	Router  = new(router)
	apiInfo = api.Api.Info
)

type router struct{ Info info }
