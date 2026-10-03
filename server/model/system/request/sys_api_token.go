package request

import (
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/system"
)

type SysApiTokenSearch struct {
	system.SysApiToken
	request.PageInfo
	Status *bool `json:"status" form:"status"`
}
