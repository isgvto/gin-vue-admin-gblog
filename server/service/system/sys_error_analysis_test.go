package system

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	"gorm.io/gorm"
)

func TestErrorAnalysisLifecycle(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	previous := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = previous; _ = sqlDB.Close() })
	if err := db.AutoMigrate(&system.SysError{}, &aiModel.AiModelConfig{}, &aiModel.ErrorAnalysisConfig{}); err != nil {
		t.Fatal(err)
	}
	var failure atomic.Bool
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
		w.Header().Set("Content-Type", "application/json")
		if failure.Load() {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(`{"error":{"message":"provider-key password=do-not-store","type":"invalid_request_error"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"错误摘要：表不存在。验证方法：确认迁移。"},"finish_reason":"stop"}]}`))
	}))
	defer server.Close()
	model := aiModel.AiModelConfig{Name: "analysis", Model: "test", Provider: "openai", BaseURL: server.URL + "/v1", APIKey: "provider-key", Status: true, MaxTokens: 512}
	if err := db.Create(&model).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&aiModel.ErrorAnalysisConfig{ID: 1, Enabled: true, ModelID: model.ID, TimeoutSeconds: 10}).Error; err != nil {
		t.Fatal(err)
	}
	source, info := "Go", "table missing"
	record := system.SysError{Form: &source, Info: &info, Status: "处理完成", AnalysisTask: "forged"}
	svc := new(SysErrorService)
	if err := svc.CreateSysError(context.Background(), &record); err != nil {
		t.Fatal(err)
	}
	if record.Status != "未处理" || record.AnalysisTask != "" {
		t.Fatal("public ingestion accepted forged analysis state")
	}
	if err := svc.GetSysErrorSolution(context.Background(), "999999"); err == nil {
		t.Fatal("accepted nonexistent log")
	}
	id := "1"
	if err := svc.GetSysErrorSolution(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	duplicateErr := svc.GetSysErrorSolution(context.Background(), id)
	close(release)
	if duplicateErr == nil {
		t.Fatal("duplicate task accepted")
	}
	wait := func(status string) system.SysError {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			var current system.SysError
			if err := db.First(&current, record.ID).Error; err != nil {
				t.Fatal(err)
			}
			if current.Status == status && current.AnalysisTask == "" {
				return current
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatalf("task did not reach %s", status)
		return system.SysError{}
	}
	success := wait("处理完成")
	if success.Solution == nil || success.AnalysisModelID != model.ID || success.AnalysisCompletedAt == nil || success.SolutionGeneratedAt == nil {
		t.Fatal("missing result metadata")
	}
	failure.Store(true)
	if err := svc.GetSysErrorSolution(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	failed := wait("处理失败")
	if failed.Solution == nil || *failed.Solution != *success.Solution || failed.SolutionModel != success.SolutionModel {
		t.Fatal("failed reanalysis lost previous result")
	}
	if failed.AnalysisError == "" || strings.Contains(failed.AnalysisError, "provider-key") || strings.Contains(failed.AnalysisError, "do-not-store") {
		t.Fatalf("unsafe failure: %s", failed.AnalysisError)
	}
	old := time.Now().Add(-10 * time.Minute)
	if err := db.Model(&record).Updates(map[string]any{"status": "处理中", "analysis_started_at": old, "analysis_task": "old-task"}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetSysError(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	recovered := wait("处理失败")
	if !strings.Contains(recovered.AnalysisError, "中断") || recovered.Solution == nil {
		t.Fatal("stale task recovery failed")
	}
	if err := db.Model(&record).Updates(map[string]any{"status": "处理中", "analysis_task": "new-task"}).Error; err != nil {
		t.Fatal(err)
	}
	runErrorAnalysis(record, aiModel.ErrorAnalysisConfig{TimeoutSeconds: 10}, model, "old-task")
	var latest system.SysError
	if err := db.First(&latest, record.ID).Error; err != nil {
		t.Fatal(err)
	}
	if latest.Status != "处理中" || latest.AnalysisTask != "new-task" {
		t.Fatal("late result overwrote newer task")
	}
	if err := db.Model(&record).Updates(map[string]any{"status": "处理失败", "analysis_task": ""}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&aiModel.ErrorAnalysisConfig{}).Where("id = ?", 1).Update("enabled", false).Error; err != nil {
		t.Fatal(err)
	}
	if err := svc.GetSysErrorSolution(context.Background(), id); err == nil {
		t.Fatal("disabled feature still submitted task")
	}
}
