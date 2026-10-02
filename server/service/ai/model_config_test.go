package ai

import (
	"encoding/json"
	"github.com/glebarez/sqlite"
	aiReq "github.com/isgvto/gin-vue-admin-gblog/server/model/ai/request"
	"gorm.io/gorm"
	"strings"
	"testing"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
)

func TestSafeModelIDContract(t *testing.T) {
	cfg := aiModel.AiModelConfig{GVA_MODEL: global.GVA_MODEL{ID: 42}, APIKey: "secret-key-1234"}
	data, err := json.Marshal(toSafeItem(cfg))
	if err != nil {
		t.Fatal(err)
	}
	var item map[string]any
	if err := json.Unmarshal(data, &item); err != nil {
		t.Fatal(err)
	}
	if item["id"] != float64(42) || item["hasKey"] != true || item["keyTail"] != "1234" {
		t.Fatalf("invalid contract: %s", data)
	}
	if strings.Contains(string(data), "secret-key") || strings.Contains(string(data), "apiKey") {
		t.Fatal("key exposed")
	}
}

func TestModelCacheUsesConfiguration(t *testing.T) {
	cfg := aiModel.AiModelConfig{APIKey: "old", Model: "model", Temperature: 0.7}
	old := modelConfigCacheKey(&cfg)
	cfg.APIKey = "new"
	if old == modelConfigCacheKey(&cfg) {
		t.Fatal("same timestamp credential update reused cache")
	}
	old = modelConfigCacheKey(&cfg)
	cfg.Temperature = 0.1
	if old == modelConfigCacheKey(&cfg) {
		t.Fatal("parameter update reused cache")
	}
}

func TestModelConfigLifecycle(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	previous := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = previous; Factory().Invalidate() })
	if err := db.AutoMigrate(&aiModel.AiModelConfig{}); err != nil {
		t.Fatal(err)
	}
	service := &ModelConfigService{}
	input := aiReq.AiModelConfigUpsert{Name: "disabled", Provider: "gemini", APIKey: "secret", Model: "test", MaxTokens: 4096}
	if err := service.Create(input); err != nil {
		t.Fatal(err)
	}
	var disabled aiModel.AiModelConfig
	if err := db.First(&disabled).Error; err != nil {
		t.Fatal(err)
	}
	if disabled.Status || disabled.Temperature != 0 || disabled.IsDefault {
		t.Fatal("zero values overwritten")
	}
	input.Name, input.Status = "enabled", true
	if err := service.Create(input); err != nil {
		t.Fatal(err)
	}
	var active aiModel.AiModelConfig
	if err := db.Where("name = ?", "enabled").First(&active).Error; err != nil {
		t.Fatal(err)
	}
	if !active.IsDefault || !active.Status || active.Temperature != 0 {
		t.Fatal("first enabled model is not default")
	}
	input.ID, input.Status = active.ID, false
	if err := service.Update(input); err == nil {
		t.Fatal("disabled default model")
	}
	if err := service.Delete(active.ID); err == nil {
		t.Fatal("deleted default model")
	}
	if err := service.SetDefault(disabled.ID); err == nil {
		t.Fatal("selected disabled model")
	}
	if err := service.Delete(0); err == nil {
		t.Fatal("accepted zero delete ID")
	}
	if err := service.SetDefault(0); err == nil {
		t.Fatal("accepted zero default ID")
	}
	input.ID, input.Name, input.Status, input.APIKey = disabled.ID, "updated", true, ""
	if err := service.Update(input); err != nil {
		t.Fatal(err)
	}
	if err := service.SetDefault(disabled.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.First(&disabled, disabled.ID).Error; err != nil {
		t.Fatal(err)
	}
	if disabled.APIKey != "secret" || !disabled.IsDefault {
		t.Fatal("edit/default failed")
	}
	input.Status, input.ID = false, active.ID
	if err := service.Update(input); err != nil {
		t.Fatal(err)
	}
	if err := service.Delete(active.ID); err != nil {
		t.Fatal(err)
	}
	var defaults int64
	db.Model(&aiModel.AiModelConfig{}).Where("is_default = ?", true).Count(&defaults)
	if defaults != 1 {
		t.Fatalf("default count = %d", defaults)
	}
	input.ID, input.MaxTokens = 0, 0
	if err := service.Create(input); err == nil {
		t.Fatal("accepted zero max tokens")
	}
}
