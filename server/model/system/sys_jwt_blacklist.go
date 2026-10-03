package system

import (
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
)

type JwtBlacklist struct {
	global.GVA_MODEL
	Jwt string `gorm:"type:text;comment:jwt"`
}
