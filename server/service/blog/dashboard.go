package blog

import (
	"time"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
)

type DashboardService struct{}

type dashboardChartItem struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

func buildChartMap(items []dashboardChartItem) map[string]interface{} {
	legend := make([]string, 0, len(items))
	series := make([]dashboardChartItem, 0, len(items))
	for _, item := range items {
		legend = append(legend, item.Name)
		series = append(series, item)
	}
	return map[string]interface{}{
		"legend": legend,
		"series": series,
	}
}

func buildVisitRecordMap(records []blogModel.VisitRecord) map[string]interface{} {
	date := make([]string, 0, len(records))
	pv := make([]int, 0, len(records))
	uv := make([]int, 0, len(records))
	for i := len(records) - 1; i >= 0; i-- {
		date = append(date, records[i].Date)
		pv = append(pv, records[i].PV)
		uv = append(uv, records[i].UV)
	}
	return map[string]interface{}{
		"date": date,
		"pv":   pv,
		"uv":   uv,
	}
}

func (s *DashboardService) GetSummary() (map[string]interface{}, error) {
	var blogCount int64
	if err := global.GVA_DB.Model(&blogModel.Blog{}).Count(&blogCount).Error; err != nil {
		return nil, err
	}

	var commentCount int64
	if err := global.GVA_DB.Model(&blogModel.Comment{}).Count(&commentCount).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)

	var todayPV int64
	if err := global.GVA_DB.Model(&blogModel.VisitLog{}).
		Where("create_time >= ? AND create_time < ?", todayStart, todayEnd).
		Where("behavior IN ?", pageViewBehaviors()).
		Count(&todayPV).Error; err != nil {
		return nil, err
	}

	var todayUV int64
	if err := global.GVA_DB.Model(&blogModel.Visitor{}).
		Where("last_time >= ? AND last_time < ?", todayStart, todayEnd).
		Count(&todayUV).Error; err != nil {
		return nil, err
	}

	var categoryItems []dashboardChartItem
	if err := global.GVA_DB.Raw(`
		SELECT c.category_name AS name, COUNT(b.id) AS value
		FROM gvto_category c
		LEFT JOIN gvto_blog b ON b.category_id = c.id
		GROUP BY c.id, c.category_name
		ORDER BY c.id
	`).Scan(&categoryItems).Error; err != nil {
		return nil, err
	}

	var tagItems []dashboardChartItem
	if err := global.GVA_DB.Raw(`
		SELECT t.tag_name AS name, COUNT(bt.blog_id) AS value
		FROM gvto_tag t
		LEFT JOIN gvto_blog_tag bt ON bt.tag_id = t.id
		GROUP BY t.id, t.tag_name
		ORDER BY t.id
	`).Scan(&tagItems).Error; err != nil {
		return nil, err
	}

	var cityVisitors []blogModel.CityVisitor
	if err := global.GVA_DB.Order("uv desc").Find(&cityVisitors).Error; err != nil {
		return nil, err
	}

	var visitRecords []blogModel.VisitRecord
	if err := global.GVA_DB.Order("id desc").Limit(30).Find(&visitRecords).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"pv":           todayPV,
		"uv":           todayUV,
		"blogCount":    blogCount,
		"commentCount": commentCount,
		"category":     buildChartMap(categoryItems),
		"tag":          buildChartMap(tagItems),
		"visitRecord":  buildVisitRecordMap(visitRecords),
		"cityVisitor":  cityVisitors,
	}, nil
}

func pageViewBehaviors() []string {
	return []string{"访问页面", "查看博客", "查看分类", "查看标签"}
}
