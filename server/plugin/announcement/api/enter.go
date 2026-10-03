package api

import "github.com/isgvto/gin-vue-admin-gblog/server/plugin/announcement/service"

var (
	Api         = new(api)
	serviceInfo = service.Service.Info
)

type api struct{ Info info }
