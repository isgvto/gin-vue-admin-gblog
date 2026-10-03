package request

import (
	"time"

	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/request"
)

type SysErrorSearch struct {
	CreatedAtRange []time.Time `json:"createdAtRange" form:"createdAtRange[]"`
	Form           *string     `json:"form" form:"form"`
	Info           *string     `json:"info" form:"info"`
	request.PageInfo
}
