package blog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	systemReq "github.com/isgvto/gin-vue-admin-gblog/server/model/system/request"
	modelService "github.com/isgvto/gin-vue-admin-gblog/server/service/ai"
	"go.uber.org/zap"
)

func TestStatusChecksConnectionOnlyWhenRequested(t *testing.T) {
	oldConfig, oldDB := global.GVA_CONFIG.AI, global.GVA_DB
	defer func() { global.GVA_CONFIG.AI, global.GVA_DB = oldConfig, oldDB; modelService.Factory().Invalidate() }()
	calls := 0
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":{"message":"invalid key","type":"invalid_api_key"}}`)
	}))
	defer provider.Close()
	global.GVA_DB = nil
	global.GVA_CONFIG.AI.Enable = true
	global.GVA_CONFIG.AI.Provider = "openai"
	global.GVA_CONFIG.AI.BaseURL = provider.URL
	global.GVA_CONFIG.AI.APIKey = "test-key"
	global.GVA_CONFIG.AI.Model = "test-model"
	modelService.Factory().Invalidate()
	for _, query := range []string{"", "?checkConnection=true"} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/blog/ai/status"+query, nil)
		ctx.Set("claims", &systemReq.CustomClaims{})
		(&AiApi{}).Status(ctx)
		var result struct {
			Data struct {
				Enabled bool   `json:"enabled"`
				Reason  string `json:"reason"`
			} `json:"data"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if query == "" && (!result.Data.Enabled || calls != 0) {
			t.Fatal("ordinary status called supplier")
		}
		if query != "" && (result.Data.Enabled || result.Data.Reason == "" || calls != 1) {
			t.Fatal("connection failure was ignored")
		}
	}
}

func TestChatValidatesBeforeQuota(t *testing.T) {
	oldLimit, oldRedis := global.GVA_CONFIG.AI.DailyLimit, global.GVA_REDIS
	defer func() { global.GVA_CONFIG.AI.DailyLimit, global.GVA_REDIS = oldLimit, oldRedis }()
	global.GVA_CONFIG.AI.DailyLimit, global.GVA_REDIS = 5, nil
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/blog/ai/chat", strings.NewReader(`{"action":"invalid"}`))
	(&AiApi{}).Chat(ctx)
	if !strings.Contains(recorder.Body.String(), "不支持的 action") {
		t.Fatal(recorder.Body.String())
	}
}

// 真实模型适配器 + Agent + API，供应商为本地模拟 HTTP 服务，不调用外部模型。
func TestChatCompletionReason(t *testing.T) {
	oldConfig, oldDB, oldLog := global.GVA_CONFIG.AI, global.GVA_DB, global.GVA_LOG
	defer func() {
		global.GVA_CONFIG.AI, global.GVA_DB, global.GVA_LOG = oldConfig, oldDB, oldLog
		modelService.Factory().Invalidate()
	}()
	global.GVA_DB, global.GVA_LOG = nil, zap.NewNop()
	for _, reason := range []string{"stop", "length", ""} {
		t.Run("reason="+reason, func(t *testing.T) {
			provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprint(w, "data: {\"id\":\"test\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":\"测试结果\"},\"finish_reason\":null}]}\n\n")
				if reason != "" {
					data, _ := json.Marshal(map[string]any{"id": "test", "object": "chat.completion.chunk", "choices": []any{map[string]any{"index": 0, "delta": map[string]any{}, "finish_reason": reason}}})
					fmt.Fprintf(w, "data: %s\n\n", data)
				}
				fmt.Fprint(w, "data: [DONE]\n\n")
			}))
			defer provider.Close()
			global.GVA_CONFIG.AI.Enable = true
			global.GVA_CONFIG.AI.Provider = "openai"
			global.GVA_CONFIG.AI.APIKey = "local-test"
			global.GVA_CONFIG.AI.BaseURL = provider.URL
			global.GVA_CONFIG.AI.Model = "local-model"
			global.GVA_CONFIG.AI.DailyLimit = 0
			modelService.Factory().Invalidate()
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/blog/ai/chat", strings.NewReader(`{"action":"custom","instruction":"测试"}`))
			(&AiApi{}).Chat(ctx)
			expected := reason
			if expected == "" {
				expected = "unknown"
			}
			body := recorder.Body.String()
			if !strings.Contains(body, `"delta":"测试结果"`) || !strings.Contains(body, `"finishReason":"`+expected+`"`) {
				t.Fatal(body)
			}
		})
	}
}

func TestSummaryUsesDistributedContext(t *testing.T) {
	oldConfig, oldDB, oldLog := global.GVA_CONFIG.AI, global.GVA_DB, global.GVA_LOG
	defer func() {
		global.GVA_CONFIG.AI, global.GVA_DB, global.GVA_LOG = oldConfig, oldDB, oldLog
		modelService.Factory().Invalidate()
	}()
	global.GVA_DB, global.GVA_LOG = nil, zap.NewNop()
	var prompt string
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
			w.WriteHeader(400)
			return
		}
		for _, msg := range body.Messages {
			prompt += msg.Content
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"test","choices":[{"index":0,"message":{"role":"assistant","content":"覆盖开头、中部和结尾的摘要"},"finish_reason":"stop"}]}`)
	}))
	defer provider.Close()
	global.GVA_CONFIG.AI.Enable = true
	global.GVA_CONFIG.AI.Provider = "openai"
	global.GVA_CONFIG.AI.APIKey = "local-test"
	global.GVA_CONFIG.AI.BaseURL = provider.URL
	global.GVA_CONFIG.AI.Model = "local-model"
	global.GVA_CONFIG.AI.DailyLimit = 0
	global.GVA_CONFIG.AI.ContextLimit = 8000
	modelService.Factory().Invalidate()
	source := "开头特征" + strings.Repeat("甲", 10000) + "中部特征" + strings.Repeat("乙", 10000) + "结尾特征"
	body, _ := json.Marshal(map[string]any{"action": "summary", "content": source, "title": "测试"})
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/blog/ai/summary", strings.NewReader(string(body)))
	(&AiApi{}).Summary(ctx)
	for _, part := range []string{"开头特征", "中部特征", "结尾特征"} {
		if !strings.Contains(prompt, part) {
			t.Fatal("missing", part)
		}
	}
	if strings.Contains(prompt, "输出永远是 Markdown") {
		t.Fatal("wrong system prompt")
	}
	if !strings.Contains(recorder.Body.String(), `"truncated":true`) || !strings.Contains(recorder.Body.String(), "覆盖开头") {
		t.Fatal(recorder.Body.String())
	}
}
