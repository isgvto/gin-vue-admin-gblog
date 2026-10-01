package response

import "github.com/isgvto/gin-vue-admin-gblog/server/config"

type SysConfigResponse struct {
	Config config.Server `json:"config"`
}
