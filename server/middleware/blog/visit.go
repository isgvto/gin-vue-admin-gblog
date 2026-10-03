package blog

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	commonResp "github.com/isgvto/gin-vue-admin-gblog/server/model/common/response"
	"github.com/isgvto/gin-vue-admin-gblog/server/utils"
	"gorm.io/gorm"
)

type VisitBehavior string

const (
	VisitBehaviorIndex         VisitBehavior = "index"
	VisitBehaviorArchive       VisitBehavior = "archive"
	VisitBehaviorMoment        VisitBehavior = "moment"
	VisitBehaviorFriend        VisitBehavior = "friend"
	VisitBehaviorAbout         VisitBehavior = "about"
	VisitBehaviorBlog          VisitBehavior = "blog"
	VisitBehaviorCategory      VisitBehavior = "category"
	VisitBehaviorTag           VisitBehavior = "tag"
	VisitBehaviorSearch        VisitBehavior = "search"
	VisitBehaviorClickFriend   VisitBehavior = "click_friend"
	VisitBehaviorLikeMoment    VisitBehavior = "like_moment"
	VisitBehaviorCheckPassword VisitBehavior = "check_password"
)

const VisitContentContextKey = "blog_visit_content"

type visitBehaviorMeta struct {
	Behavior string
	Content  string
	Remark   string
	CountPV  bool
}

func SetVisitContent(c *gin.Context, content string) {
	c.Set(VisitContentContextKey, strings.TrimSpace(content))
}

func VisitRecord(behavior VisitBehavior) gin.HandlerFunc {
	return func(c *gin.Context) {
		identification := c.GetHeader("identification")
		if identification == "" {
			identification = uuid.NewString()
			c.Header("identification", identification)
			c.Header("Access-Control-Expose-Headers", "identification")
		}

		now := time.Now()
		ip := utils.BlogClientIP(c.Request)
		ipSource := utils.BlogIPSource(ip)
		ua := c.Request.UserAgent()
		os, browser := utils.BlogParseUserAgent(ua)
		dateKey := now.Format("2006-01-02")
		isNewVisitor := false
		var visitor blogModel.Visitor
		err := global.GVA_DB.Where("uuid = ?", identification).First(&visitor).Error
		if err == nil {
			updates := map[string]interface{}{
				"last_time": now,
				"ip":        ip,
				"ip_source": ipSource,
				"os":        os,
				"browser":   browser,
			}
			_ = global.GVA_DB.Model(&blogModel.Visitor{}).Where("id = ?", visitor.ID).Updates(updates).Error
		} else if err == gorm.ErrRecordNotFound {
			isNewVisitor = true
			_ = global.GVA_DB.Create(&blogModel.Visitor{
				UUID:       identification,
				IP:         &ip,
				IPSource:   &ipSource,
				OS:         &os,
				Browser:    &browser,
				CreateTime: now,
				LastTime:   now,
				PV:         intPtr(0),
				UserAgent:  &ua,
			}).Error
			var city blogModel.CityVisitor
			if e := global.GVA_DB.Where("city = ?", ipSource).First(&city).Error; e == nil {
				_ = global.GVA_DB.Model(&blogModel.CityVisitor{}).Where("city = ?", ipSource).UpdateColumn("uv", gorm.Expr("uv + 1")).Error
			} else if e == gorm.ErrRecordNotFound {
				_ = global.GVA_DB.Create(&blogModel.CityVisitor{City: ipSource, UV: 1}).Error
			}
		}

		var body []byte
		if c.Request.Method != "GET" && c.Request.Body != nil {
			body, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		start := time.Now()
		c.Next()
		cost := time.Since(start)

		param := c.Request.URL.RawQuery
		if len(body) > 0 {
			param = string(body)
		}
		success := isVisitResponseSuccess(c)
		meta := buildVisitBehaviorMeta(behavior, c, body, success)
		if success && meta.CountPV {
			increaseVisitorPV(identification)
			increaseDailyVisit(dateKey, isNewVisitor || isFirstPageViewToday(identification, now))
		}
		_ = global.GVA_DB.Create(&blogModel.VisitLog{
			UUID:       &identification,
			URI:        c.Request.URL.Path,
			Method:     c.Request.Method,
			Param:      param,
			Behavior:   &meta.Behavior,
			Content:    stringPtr(meta.Content),
			Remark:     stringPtr(meta.Remark),
			IP:         &ip,
			IPSource:   &ipSource,
			OS:         &os,
			Browser:    &browser,
			Times:      int(cost.Milliseconds()),
			CreateTime: time.Now(),
			UserAgent:  &ua,
		}).Error
	}
}

func buildVisitBehaviorMeta(behavior VisitBehavior, c *gin.Context, body []byte, success bool) visitBehaviorMeta {
	page := firstNonEmpty(c.Query("page"), c.Query("pageNum"), "1")
	meta := visitBehaviorDefaults(behavior)
	switch behavior {
	case VisitBehaviorIndex, VisitBehaviorMoment:
		meta.Remark = "第" + page + "页"
	case VisitBehaviorBlog:
		if success {
			if title := visitContent(c); title != "" {
				meta.Content = title
				meta.Remark = "文章标题：" + title
			}
		}
	case VisitBehaviorSearch:
		if success {
			query := strings.TrimSpace(c.Query("query"))
			meta.Content = query
			meta.Remark = "搜索内容：" + query
		}
	case VisitBehaviorCategory:
		categoryName := strings.TrimSpace(c.Query("categoryName"))
		meta.Content = categoryName
		meta.Remark = "分类名称：" + categoryName + "，第" + page + "页"
	case VisitBehaviorTag:
		tagName := strings.TrimSpace(c.Query("tagName"))
		meta.Content = tagName
		meta.Remark = "标签名称：" + tagName + "，第" + page + "页"
	case VisitBehaviorClickFriend:
		nickname := firstNonEmpty(c.Query("nickname"), c.PostForm("nickname"), requestValue(body, "nickname"))
		meta.Content = nickname
		meta.Remark = "友链名称：" + nickname
	}
	return meta
}

func visitBehaviorDefaults(behavior VisitBehavior) visitBehaviorMeta {
	switch behavior {
	case VisitBehaviorIndex:
		return visitBehaviorMeta{Behavior: "访问页面", Content: "首页", CountPV: true}
	case VisitBehaviorArchive:
		return visitBehaviorMeta{Behavior: "访问页面", Content: "归档", CountPV: true}
	case VisitBehaviorMoment:
		return visitBehaviorMeta{Behavior: "访问页面", Content: "动态", CountPV: true}
	case VisitBehaviorFriend:
		return visitBehaviorMeta{Behavior: "访问页面", Content: "友链", CountPV: true}
	case VisitBehaviorAbout:
		return visitBehaviorMeta{Behavior: "访问页面", Content: "关于我", CountPV: true}
	case VisitBehaviorBlog:
		return visitBehaviorMeta{Behavior: "查看博客", CountPV: true}
	case VisitBehaviorCategory:
		return visitBehaviorMeta{Behavior: "查看分类", CountPV: true}
	case VisitBehaviorTag:
		return visitBehaviorMeta{Behavior: "查看标签", CountPV: true}
	case VisitBehaviorSearch:
		return visitBehaviorMeta{Behavior: "搜索博客"}
	case VisitBehaviorClickFriend:
		return visitBehaviorMeta{Behavior: "点击友链"}
	case VisitBehaviorLikeMoment:
		return visitBehaviorMeta{Behavior: "点赞动态"}
	case VisitBehaviorCheckPassword:
		return visitBehaviorMeta{Behavior: "校验博客密码"}
	default:
		return visitBehaviorMeta{Behavior: "UNKNOWN", Content: "UNKNOWN"}
	}
}

func isVisitResponseSuccess(c *gin.Context) bool {
	if code, ok := c.Get(commonResp.ResponseCodeContextKey); ok {
		if value, ok := code.(int); ok {
			return value == commonResp.SUCCESS
		}
	}
	status := c.Writer.Status()
	return status >= 200 && status < 300
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func requestValue(body []byte, key string) string {
	if len(body) == 0 {
		return ""
	}
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err == nil {
		if value, ok := obj[key]; ok {
			return strings.TrimSpace(toString(value))
		}
	}
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(values.Get(key))
}

func toString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return ""
	}
}

func visitContent(c *gin.Context) string {
	value, ok := c.Get(VisitContentContextKey)
	if !ok {
		return ""
	}
	content, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(content)
}

func isFirstPageViewToday(uuid string, now time.Time) bool {
	if uuid == "" {
		return true
	}
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)
	var total int64
	err := global.GVA_DB.Model(&blogModel.VisitLog{}).
		Where("uuid = ? AND create_time >= ? AND create_time < ?", uuid, start, end).
		Where("behavior IN ?", []string{"访问页面", "查看博客", "查看分类", "查看标签"}).
		Count(&total).Error
	return err != nil || total == 0
}

func stringPtr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func increaseVisitorPV(uuid string) {
	if uuid == "" {
		return
	}
	_ = global.GVA_DB.Model(&blogModel.Visitor{}).
		Where("uuid = ?", uuid).
		UpdateColumn("pv", gorm.Expr("coalesce(pv,0) + 1")).Error
}

func intPtr(v int) *int {
	return &v
}

func increaseDailyVisit(dateKey string, increaseUV bool) {
	var visitRecord blogModel.VisitRecord
	err := global.GVA_DB.Where("date = ?", dateKey).First(&visitRecord).Error
	if err == nil {
		updates := map[string]interface{}{
			"pv": gorm.Expr("pv + 1"),
		}
		if increaseUV {
			updates["uv"] = gorm.Expr("uv + 1")
		}
		_ = global.GVA_DB.Model(&blogModel.VisitRecord{}).Where("id = ?", visitRecord.ID).Updates(updates).Error
		return
	}
	if err == gorm.ErrRecordNotFound {
		record := blogModel.VisitRecord{Date: dateKey, PV: 1}
		if increaseUV {
			record.UV = 1
		}
		_ = global.GVA_DB.Create(&record).Error
	}
}
