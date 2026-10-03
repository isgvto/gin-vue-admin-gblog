package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/glebarez/sqlite"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	aiReq "github.com/isgvto/gin-vue-admin-gblog/server/model/ai/request"
	"gorm.io/gorm"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndependentImageConfigAndKeyPreservation(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	old := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = old })
	if err = db.AutoMigrate(&aiModel.ImageConfig{}); err != nil {
		t.Fatal(err)
	}
	service := &ImageService{}
	input := ImageConfigInput{Enabled: true, Provider: "ark", Model: "image-model", APIKey: "secret", TimeoutSeconds: 180}
	if err = service.Save(input); err != nil {
		t.Fatal(err)
	}
	cfg, _ := service.Config()
	endpoint, err := service.Resolve(cfg)
	if err != nil || endpoint.BaseURL != defaultArkBaseURL || endpoint.APIKey != "secret" {
		t.Fatalf("independent resolve: %v", err)
	}
	input.APIKey = ""
	input.Model = "changed"
	if err = service.Save(input); err != nil {
		t.Fatal(err)
	}
	cfg, _ = service.Config()
	if cfg.APIKey != "secret" || cfg.Model != "changed" {
		t.Fatal("key retention failed")
	}
	raw, _ := json.Marshal(cfg)
	if strings.Contains(string(raw), "secret") || strings.Contains(string(raw), "modelId") {
		t.Fatal("secret or binding exposed")
	}
	input.ClearKey = true
	if err = service.Save(input); err == nil {
		t.Fatal("enabled config accepted empty key")
	}
	input.Enabled = false
	if err = service.Save(input); err != nil {
		t.Fatal(err)
	}
	cfg, _ = service.Config()
	if cfg.APIKey != "" {
		t.Fatal("key not cleared")
	}
}

func TestExplicitImageProbeUsesImageEndpoint(t *testing.T) {
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, 8, 8))); err != nil {
		t.Fatal(err)
	}
	calls := 0
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/v3/images/generations" {
			t.Errorf("image test used %s", r.URL.Path)
			w.WriteHeader(404)
			return
		}
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Error("missing shared key")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["size"] != "2048x2048" || body["response_format"] != "b64_json" || body["model"] != "user-selected-model" {
			t.Errorf("wrong image payload: %v", body)
		}
		if _, exists := body["messages"]; exists {
			t.Error("chat messages sent to image endpoint")
		}
		if _, exists := body["n"]; exists {
			t.Error("OpenAI n sent to Ark")
		}
		if _, exists := body["sequential_image_generation"]; exists {
			t.Error("unsupported group option sent to Flash/Pro")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{map[string]string{"b64_json": base64.StdEncoding.EncodeToString(buffer.Bytes())}}})
	}))
	defer provider.Close()
	req := aiReq.AiModelConfigUpsert{Provider: "ark", BaseURL: provider.URL + "/v3/", APIKey: "secret", Model: "user-selected-model", TestMode: "image"}
	if err := (&ModelConfigService{}).TestConnection(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("image probe did not generate exactly once")
	}
	req.TestMode = "unknown"
	if err := (&ModelConfigService{}).TestConnection(context.Background(), req); err == nil || calls != 1 {
		t.Fatal("invalid test mode dispatched")
	}
}

func TestImageRateLimitMessage(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		_, _ = w.Write([]byte(`{"error":{"message":"secret"}}`))
	}))
	defer provider.Close()
	_, err := (&ImageService{}).GenerateWithConfig(context.Background(), ImageEndpointConfig{BaseURL: provider.URL, Model: "user-model", APIKey: "secret", TimeoutSeconds: 10}, "测试配图", "1024x1024")
	if err == nil || !strings.Contains(err.Error(), "限流") || !strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "secret") {
		t.Fatalf("wrong rate limit message: %v", err)
	}
}

func TestImageBadRequestDiagnostics(t *testing.T) {
	for _, test := range []struct{ name, body, expected string }{
		{"nested", `{"error":{"code":"InvalidParameter","message":"size must be at least 3686400 pixels; api_key=secret","param":"size"}}`, "InvalidParameter"},
		{"top-level", `{"code":"InvalidModel","message":"unsupported model"}`, "unsupported model"},
		{"html", `<html>secret arbitrary server debug</html>`, "HTTP 400"},
	} {
		t.Run(test.name, func(t *testing.T) {
			provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(400); _, _ = w.Write([]byte(test.body)) }))
			defer provider.Close()
			_, err := (&ImageService{}).GenerateWithConfig(context.Background(), ImageEndpointConfig{Provider: "ark", BaseURL: provider.URL, Model: "selected-model", APIKey: "secret", TimeoutSeconds: 10}, "测试配图", "1536x1024")
			if err == nil || !strings.Contains(err.Error(), test.expected) || !strings.Contains(err.Error(), "2496x1664") || strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "<html>") {
				t.Fatalf("unsafe/incomplete diagnostics: %v", err)
			}
		})
	}
}
