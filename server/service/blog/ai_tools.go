package blog

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	"go.uber.org/zap"
)

// 写作助手可调用的只读工具，全部复用既有 blog service，不写任何新 SQL。

var (
	aiArticleSvc      = &ArticleService{}
	aiAdminArticleSvc = &AdminArticleService{}
)

type searchBlogsParams struct {
	Query string `json:"query" jsonschema:"description=搜索博主历史文章的关键词"`
	Limit int    `json:"limit,omitempty" jsonschema:"description=返回条数，默认3，最大5"`
}

type blogContentParams struct {
	ID uint `json:"id" jsonschema:"description=search_my_blogs 返回的文章 ID"`
}

func toolInferSearchBlogs() (tool.InvokableTool, error) {
	return toolutils.InferTool(
		"search_my_blogs",
		"按关键词搜索博主已发布的历史文章，返回标题、摘要与ID，用于学习博主的行文风格",
		func(ctx context.Context, p searchBlogsParams) (string, error) {
			limit := p.Limit
			if limit <= 0 || limit > aiToolBlogLimit {
				limit = 3
			}
			items, err := aiArticleSvc.SearchPublishedBlogs(p.Query)
			if err != nil {
				global.GVA_LOG.Error("AI 工具搜索文章失败", zap.Error(err))
				return "搜索失败，请基于已有信息作答", nil
			}
			if len(items) > limit {
				items = items[:limit]
			}
			if len(items) == 0 {
				return "没有搜索到相关文章", nil
			}
			out := ""
			for _, item := range items {
				content := []rune(item.Content)
				if len(content) > 400 {
					content = content[:400]
				}
				out += fmt.Sprintf("【ID:%d】%s\n%s\n\n", item.ID, item.Title, string(content))
			}
			return out, nil
		})
}

func toolInferBlogContent() (tool.InvokableTool, error) {
	return toolutils.InferTool(
		"get_blog_content",
		"获取博主某篇历史文章的正文内容（截断），用于深入参考特定文章的写法",
		func(ctx context.Context, p blogContentParams) (string, error) {
			if p.ID == 0 {
				return "无效的文章ID", nil
			}
			blog, err := aiArticleSvc.GetPublishedByID(p.ID)
			if err != nil {
				global.GVA_LOG.Error("AI 工具读取文章失败", zap.Error(err))
				return "读取文章失败，请基于已有信息作答", nil
			}
			content := []rune(blog.Content)
			if len(content) > aiToolBodyLimit {
				content = content[:aiToolBodyLimit]
			}
			return fmt.Sprintf("《%s》\n\n%s", blog.Title, string(content)), nil
		})
}
