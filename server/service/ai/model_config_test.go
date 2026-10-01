package ai

import (
	"encoding/json"
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
