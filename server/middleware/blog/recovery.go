package blog

import (
	"fmt"
	"net/http/httputil"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	"github.com/isgvto/gin-vue-admin-gblog/server/utils"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				req, _ := httputil.DumpRequest(c.Request, false)
				ip := utils.BlogClientIP(c.Request)
				ipSource := utils.BlogIPSource(ip)
				ua := c.Request.UserAgent()
				os, browser := utils.BlogParseUserAgent(ua)
				desc := "blog panic"
				errMsg := fmt.Sprintf("panic: %v\nrequest: %s\nstack: %s", err, string(req), string(debug.Stack()))
				_ = global.GVA_DB.Create(&blogModel.ExceptionLog{
					URI:         c.Request.URL.Path,
					Method:      c.Request.Method,
					Param:       strPtr(limitString(c.Request.URL.RawQuery, 2000)),
					Description: &desc,
					Error:       &errMsg,
					IP:          &ip,
					IPSource:    &ipSource,
					OS:          &os,
					Browser:     &browser,
					CreateTime:  time.Now(),
					UserAgent:   &ua,
				}).Error
				panic(err)
			}
		}()
		c.Next()
	}
}
