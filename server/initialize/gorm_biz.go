package initialize

import (
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
)

func bizModel() error {
	db := global.GVA_DB
	err := db.AutoMigrate(
		aiModel.AiModelConfig{},
		aiModel.ErrorAnalysisConfig{},
		aiModel.WorkflowConfig{},
		aiModel.ImageConfig{},
		blogModel.AiVisualTask{},
		blogModel.About{},
		blogModel.Blog{},
		blogModel.BlogTag{},
		blogModel.Category{},
		blogModel.CityVisitor{},
		blogModel.Comment{},
		blogModel.ExceptionLog{},
		blogModel.Friend{},
		blogModel.Moment{},
		blogModel.OperationLog{},
		blogModel.ScheduleJobLog{},
		blogModel.SiteSetting{},
		blogModel.Tag{},
		blogModel.VisitLog{},
		blogModel.VisitRecord{},
		blogModel.Visitor{},
	)
	if err != nil {
		return err
	}
	if err := migrateImageModelBinding(db); err != nil {
		return err
	}
	if err := migrateErrorAnalysisAccess(db); err != nil {
		return err
	}
	if err := migrateWorkflowAccess(db); err != nil {
		return err
	}
	if err := migrateVisualAccess(db); err != nil {
		return err
	}
	return migrateGitHubProfileAccess(db)
}
