package response

import blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"

type Option struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

type CategoryAndTagResponse struct {
	Categories []blogModel.Category `json:"categories"`
	Tags       []blogModel.Tag      `json:"tags"`
}

type SiteStats struct {
	ArticleCount  int64 `json:"articleCount"`
	CategoryCount int64 `json:"categoryCount"`
	TagCount      int64 `json:"tagCount"`
}
