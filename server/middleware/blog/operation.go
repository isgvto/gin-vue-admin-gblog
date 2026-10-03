package blog

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	"github.com/isgvto/gin-vue-admin-gblog/server/utils"
)

func OperationRecord() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body []byte
		if c.Request.Method != http.MethodGet {
			body, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		} else {
			query, _ := url.QueryUnescape(c.Request.URL.RawQuery)
			body = []byte(query)
		}

		username := "anonymous"
		if claims, _ := utils.GetClaims(c); claims != nil && claims.Username != "" {
			username = claims.Username
		}

		writer := responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer
		start := time.Now()

		c.Next()

		ip := utils.BlogClientIP(c.Request)
		ipSource := utils.BlogIPSource(ip)
		ua := c.Request.UserAgent()
		os, browser := utils.BlogParseUserAgent(ua)
		record := blogModel.OperationLog{
			Username:    username,
			URI:         c.Request.URL.Path,
			Method:      c.Request.Method,
			Param:       strPtr(limitString(string(body), 2000)),
			IP:          &ip,
			IPSource:    &ipSource,
			OS:          &os,
			Browser:     &browser,
			CreateTime:  time.Now(),
			UserAgent:   &ua,
			Times:       int(time.Since(start).Milliseconds()),
			Description: strPtr(c.Request.URL.Path),
		}
		if errStr := strings.TrimSpace(c.Errors.ByType(gin.ErrorTypePrivate).String()); errStr != "" {
			record.Description = strPtr(errStr)
		}
		_ = global.GVA_DB.Create(&record).Error
	}
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func strPtr(v string) *string {
	return &v
}

func limitString(v string, n int) string {
	if len(v) <= n {
		return v
	}
	return v[:n]
}
