package request

import (
	"time"

	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/request"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/system"
)

type SysExportTemplateSearch struct {
	system.SysExportTemplate
	StartCreatedAt *time.Time `json:"startCreatedAt" form:"startCreatedAt"`
	EndCreatedAt   *time.Time `json:"endCreatedAt" form:"endCreatedAt"`
	request.PageInfo
}
