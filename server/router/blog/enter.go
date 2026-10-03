package blog

import api "github.com/isgvto/gin-vue-admin-gblog/server/api/v1"

type RouterGroup struct {
	SiteRouter
	ArchiveRouter
	CategoryRouter
	TagRouter
	ArticleRouter
	AdminArticleRouter
	AboutRouter
	FriendRouter
	MomentRouter
	CommentRouter
	SiteSettingRouter
	DashboardRouter
	VisitLogRouter
	VisitorRouter
	ExceptionLogRouter
	OperationLogRouter
	TelegramRouter
	AuthRouter
	DocsRouter
	AiRouter
}

var (
	siteApi         = api.ApiGroupApp.BlogApiGroup.SiteApi
	archiveApi      = api.ApiGroupApp.BlogApiGroup.ArchiveApi
	categoryApi     = api.ApiGroupApp.BlogApiGroup.CategoryApi
	tagApi          = api.ApiGroupApp.BlogApiGroup.TagApi
	articleApi      = api.ApiGroupApp.BlogApiGroup.ArticleApi
	adminArticleApi = api.ApiGroupApp.BlogApiGroup.AdminArticleApi
	aboutApi        = api.ApiGroupApp.BlogApiGroup.AboutApi
	friendApi       = api.ApiGroupApp.BlogApiGroup.FriendApi
	momentApi       = api.ApiGroupApp.BlogApiGroup.MomentApi
	commentApi      = api.ApiGroupApp.BlogApiGroup.CommentApi
	siteSettingApi  = api.ApiGroupApp.BlogApiGroup.SiteSettingApi
	dashboardApi    = api.ApiGroupApp.BlogApiGroup.DashboardApi
	visitLogApi     = api.ApiGroupApp.BlogApiGroup.VisitLogApi
	visitorApi      = api.ApiGroupApp.BlogApiGroup.VisitorApi
	exceptionLogApi = api.ApiGroupApp.BlogApiGroup.ExceptionLogApi
	operationLogApi = api.ApiGroupApp.BlogApiGroup.OperationLogApi
	telegramApi     = api.ApiGroupApp.BlogApiGroup.TelegramApi
	authApi         = api.ApiGroupApp.BlogApiGroup.AuthApi
	docsApi         = api.ApiGroupApp.BlogApiGroup.DocsApi
	aiApi           = api.ApiGroupApp.BlogApiGroup.AiApi
)
