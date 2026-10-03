package response

import (
	"github.com/isgvto/gin-vue-admin-gblog/server/model/system/request"
)

type PolicyPathResponse struct {
	Paths []request.CasbinInfo `json:"paths"`
}
