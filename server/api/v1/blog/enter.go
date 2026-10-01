package blog

import "github.com/isgvto/gin-vue-admin-gblog/server/service"

type ApiGroup struct {
	SiteApi         SiteApi
	ArchiveApi      ArchiveApi
	CategoryApi     CategoryApi
	TagApi          TagApi
	ArticleApi      ArticleApi
	AdminArticleApi AdminArticleApi
	AboutApi        AboutApi
	FriendApi       FriendApi
	MomentApi       MomentApi
	CommentApi      CommentApi
	SiteSettingApi  SiteSettingApi
	DashboardApi    DashboardApi
	VisitLogApi     VisitLogApi
	VisitorApi      VisitorApi
	ExceptionLogApi ExceptionLogApi
	OperationLogApi OperationLogApi
	TelegramApi     TelegramApi
	AuthApi         AuthApi
	DocsApi         DocsApi
	AiApi           AiApi
}

var (
	siteService         = service.ServiceGroupApp.BlogServiceGroup.SiteService
	archiveService      = service.ServiceGroupApp.BlogServiceGroup.ArchiveService
	categoryService     = service.ServiceGroupApp.BlogServiceGroup.CategoryService
	tagService          = service.ServiceGroupApp.BlogServiceGroup.TagService
	articleService      = service.ServiceGroupApp.BlogServiceGroup.ArticleService
	adminArticleService = service.ServiceGroupApp.BlogServiceGroup.AdminArticleService
	aboutService        = service.ServiceGroupApp.BlogServiceGroup.AboutService
	friendService       = service.ServiceGroupApp.BlogServiceGroup.FriendService
	momentService       = service.ServiceGroupApp.BlogServiceGroup.MomentService
	commentService      = service.ServiceGroupApp.BlogServiceGroup.CommentService
	siteSettingService  = service.ServiceGroupApp.BlogServiceGroup.SiteSettingService
	dashboardService    = service.ServiceGroupApp.BlogServiceGroup.DashboardService
	visitLogService     = service.ServiceGroupApp.BlogServiceGroup.VisitLogService
	visitorService      = service.ServiceGroupApp.BlogServiceGroup.VisitorService
	exceptionLogService = service.ServiceGroupApp.BlogServiceGroup.ExceptionLogService
	operationLogService = service.ServiceGroupApp.BlogServiceGroup.OperationLogService
	telegramService     = service.ServiceGroupApp.BlogServiceGroup.TelegramService
	authService         = service.ServiceGroupApp.BlogServiceGroup.AuthService
	docsService         = service.ServiceGroupApp.BlogServiceGroup.DocsService
	aiService           = service.ServiceGroupApp.BlogServiceGroup.AiService
)
