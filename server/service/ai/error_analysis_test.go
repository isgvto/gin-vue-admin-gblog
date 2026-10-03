package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	"gorm.io/gorm"
)

func TestErrorAnalysisBinding(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	previous := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = previous })
	if err := db.AutoMigrate(&aiModel.AiModelConfig{}, &aiModel.ErrorAnalysisConfig{}); err != nil {
		t.Fatal(err)
	}
	svc := ErrorAnalysisService{}
	cfg, err := svc.Config()
	if err != nil || cfg.Enabled || cfg.TimeoutSeconds != 60 {
		t.Fatalf("invalid initial config: %+v %v", cfg, err)
	}
	if err := svc.Save(aiModel.ErrorAnalysisConfig{Enabled: true, TimeoutSeconds: 60}); err == nil {
		t.Fatal("accepted missing default")
	}
	model := aiModel.AiModelConfig{Name: "dedicated", Model: "test", Provider: "openai", Status: true, APIKey: "key"}
	if err := db.Create(&model).Error; err != nil {
		t.Fatal(err)
	}
	cfg = aiModel.ErrorAnalysisConfig{Enabled: true, ModelID: model.ID, TimeoutSeconds: 60}
	if err := svc.Save(cfg); err != nil {
		t.Fatal(err)
	}
	if err := ensureNotBoundToErrorAnalysis(model.ID, false); err == nil {
		t.Fatal("allowed deletion of bound model")
	}
	if err := (new(ModelConfigService)).Delete(model.ID); err == nil {
		t.Fatal("delete bypassed binding")
	}
	if err := db.Model(&aiModel.ErrorAnalysisConfig{}).Where("id = ?", 1).Update("permissions_migrated", true).Error; err != nil {
		t.Fatal(err)
	}
	cfg.Enabled = false
	if err := svc.Save(cfg); err != nil {
		t.Fatal(err)
	}
	saved, err := svc.Config()
	if err != nil || !saved.PermissionsMigrated || saved.Enabled {
		t.Fatal("zero-value save or migration marker lost")
	}
	if err := db.Model(&model).Update("status", false).Error; err != nil {
		t.Fatal(err)
	}
	other := aiModel.AiModelConfig{Name: "default", Model: "other", Status: true, IsDefault: true, APIKey: "other-key"}
	if err := db.Create(&other).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Resolve(cfg); err == nil {
		t.Fatal("silently replaced disabled explicit model")
	}
	resolved, err := svc.Resolve(aiModel.ErrorAnalysisConfig{TimeoutSeconds: 60})
	if err != nil || resolved.ID != other.ID {
		t.Fatalf("explicit default selection failed: %v", err)
	}
	if err := svc.Save(aiModel.ErrorAnalysisConfig{TimeoutSeconds: 0}); err == nil {
		t.Fatal("accepted zero timeout")
	}
}

func TestErrorAnalysisRedaction(t *testing.T) {
	input := "password=hunter2\nAuthorization: Bearer sensitive.jwt.token\n{\"api_key\":\"opaque-secret\"}\nCookie: session=secret-session; csrf=secret-csrf\nhttps://user:db-pass@example.com\nsk-abcdefghijk"
	output := SanitizeAnalysisText(input)
	for _, secret := range []string{"hunter2", "sensitive.jwt.token", "opaque-secret", "secret-session", "secret-csrf", "db-pass", "sk-abcdefghijk"} {
		if strings.Contains(output, secret) {
			t.Fatalf("redaction leaked %s", secret)
		}
	}
	if !strings.Contains(SanitizeAnalysisText(strings.Repeat("中", 13000)), "已截断") {
		t.Fatal("missing unicode-safe limit")
	}
}

func TestErrorAnalysisRealAdapter(t *testing.T) {
	var submitted string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		for _, message := range body.Messages {
			submitted += message.Content
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"错误摘要：数据库表不存在。排查步骤：检查迁移。"},"finish_reason":"stop"}]}`))
	}))
	defer server.Close()
	model := &aiModel.AiModelConfig{Provider: "openai", BaseURL: server.URL + "/v1", APIKey: "opaque-model-key", Model: "test", MaxTokens: 512}
	result, err := (ErrorAnalysisService{}).Analyze(context.Background(), aiModel.ErrorAnalysisConfig{TimeoutSeconds: 10}, model, "Go", "password=supersecret opaque-model-key")
	if err != nil || !strings.Contains(result, "数据库表") {
		t.Fatalf("adapter failed: %s %v", result, err)
	}
	if strings.Contains(submitted, "supersecret") || strings.Contains(submitted, model.APIKey) {
		t.Fatal("sent sensitive log to provider")
	}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	if _, err := (ErrorAnalysisService{}).Analyze(ctx, aiModel.ErrorAnalysisConfig{TimeoutSeconds: 10}, model, "Go", "error"); err == nil || !strings.Contains(err.Error(), "超时") {
		t.Fatalf("deadline not enforced: %v", err)
	}
}
