package request

import (
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/system"
)

type SysLoginLogSearch struct {
	system.SysLoginLog
	request.PageInfo
}
