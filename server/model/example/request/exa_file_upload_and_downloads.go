package request

import (
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/request"
)

type ExaAttachmentCategorySearch struct {
	ClassId int `json:"classId" form:"classId"`
	request.PageInfo
}
