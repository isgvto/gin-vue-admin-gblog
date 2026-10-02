package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	aiReq "github.com/isgvto/gin-vue-admin-gblog/server/model/ai/request"
)

func TestProviderProbe(t *testing.T) {
	called := 0
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-secret" {
			t.Error("missing auth")
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/models":
			fmt.Fprint(w, `{"data":[{"id":"z-model"},{"id":"a-model"},{"id":"a-model"}]}`)
		case "/v1/chat/completions":
			called++
			var payload struct {
				Model       string  `json:"model"`
				Temperature float32 `json:"temperature"`
				MaxTokens   int     `json:"max_tokens"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Error(err)
			}
			if payload.Model != "a-model" || payload.Temperature != 0 || payload.MaxTokens != 4096 {
				t.Errorf("wrong parameters: %+v", payload)
			}
			fmt.Fprint(w, `{"id":"test","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"OK"},"finish_reason":"stop"}]}`)
		default:
			t.Error(r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer provider.Close()
	service := &ModelConfigService{}
	req := aiReq.AiModelConfigUpsert{Provider: "openai", BaseURL: provider.URL + "/v1/", APIKey: "test-secret", Model: "a-model", MaxTokens: 4096}
	models, err := service.ListProviderModels(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 2 || models[0].ID != "a-model" {
		t.Fatalf("models = %+v", models)
	}
	req.BaseURL = provider.URL + "/v1"
	if err := service.TestConnection(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if called != 1 {
		t.Fatal("probe did not perform generation")
	}
}

func TestProbeFailureAndCancellation(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":{"message":"bad key test-secret","type":"invalid_api_key"}}`)
	}))
	defer provider.Close()
	service := &ModelConfigService{}
	req := aiReq.AiModelConfigUpsert{Provider: "openai", BaseURL: provider.URL, APIKey: "test-secret", Model: "model", MaxTokens: 4096}
	if err := service.TestConnection(context.Background(), req); err == nil || strings.Contains(err.Error(), "test-secret") {
		t.Fatalf("unsafe or missing error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := service.TestConnection(ctx, req); err == nil {
		t.Fatal("ignored cancellation")
	}
	if _, err := service.ListProviderModels(context.Background(), req); err == nil {
		t.Fatal("ignored list auth error")
	}
}

func TestUnsupportedModelList(t *testing.T) {
	provider := httptest.NewServer(http.NotFoundHandler())
	defer provider.Close()
	_, err := (&ModelConfigService{}).ListProviderModels(context.Background(), aiReq.AiModelConfigUpsert{Provider: "ark", BaseURL: provider.URL, APIKey: "key"})
	if err == nil || !strings.Contains(err.Error(), "手动填写") {
		t.Fatal(err)
	}
}
