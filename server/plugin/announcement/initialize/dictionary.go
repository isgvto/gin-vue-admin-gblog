package initialize

import (
	"context"

	model "github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	"github.com/isgvto/gin-vue-admin-gblog/server/plugin/plugin-tool/utils"
)

func Dictionary(ctx context.Context) {
	entities := []model.SysDictionary{}
	utils.RegisterDictionaries(entities...)
}
