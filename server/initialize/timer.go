package initialize

import (
	"fmt"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	"github.com/isgvto/gin-vue-admin-gblog/server/service"
	"github.com/isgvto/gin-vue-admin-gblog/server/task"
	"github.com/robfig/cron/v3"
)

const blogViewsSyncJobID uint = 1

func Timer() {
	go func() {
		var option []cron.Option
		option = append(option, cron.WithSeconds())

		_, err := global.GVA_Timer.AddTaskByFunc("ClearDB", "@daily", func() {
			err := task.ClearTable(global.GVA_DB)
			if err != nil {
				fmt.Println("timer error:", err)
			}
		}, "定时清理数据库【日志，黑名单】内容", option...)
		if err != nil {
			fmt.Println("add timer error:", err)
		}

		_, err = global.GVA_Timer.AddTaskByFunc("BlogViewsSync", "0 0 1 * * *", func() {
			err := task.RecordScheduleJobLog(
				blogViewsSyncJobID,
				"blogArticleService",
				"SyncViewsToDatabase",
				"",
				service.ServiceGroupApp.BlogServiceGroup.ArticleService.SyncViewsToDatabase,
			)
			if err != nil {
				fmt.Println("timer error:", err)
			}
		}, "每天凌晨一点同步博客文章浏览量到数据库", option...)
		if err != nil {
			fmt.Println("add timer error:", err)
		}
	}()
}
