package system

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	model "github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	systemReq "github.com/isgvto/gin-vue-admin-gblog/server/model/system/request"
	"gorm.io/gorm"
)

func TestWorkflowChatStreamingAndBlocking(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	previous := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = previous })
	if err := db.AutoMigrate(&aiModel.WorkflowConfig{}, &aiModel.AiModelConfig{}, &model.SysAIWorkflowSession{}); err != nil {
		t.Fatal(err)
	}
	var invalid bool
	var submitted string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Stream   bool `json:"stream"`
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		for _, m := range body.Messages {
			submitted += m.Content
		}
		text := `{"summary":"分类需求 opaque-credential","modules":[{"name":"category","fields":[]}]}`
		if len(body.Messages) > 0 && strings.Contains(body.Messages[0].Content, `"steps"`) {
			text = `{"summary":"开发路线","steps":[{"title":"预览","prompt":"检查字段并预览代码","autoExecutable":true}]}`
		}
		if invalid {
			text = "not valid JSON"
		}
		if !body.Stream {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"role": "assistant", "content": text}, "finish_reason": "stop"}}})
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		runes := []rune(text)
		for i := 0; i < len(runes); i += 3 {
			end := i + 3
			if end > len(runes) {
				end = len(runes)
			}
			chunk, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"index": 0, "delta": map[string]any{"content": string(runes[i:end])}}}})
			_, _ = fmt.Fprintf(w, "data: %s\n\n", chunk)
			w.(http.Flusher).Flush()
		}
		_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer upstream.Close()
	config := aiModel.AiModelConfig{Name: "dedicated", Provider: "openai", BaseURL: upstream.URL + "/v1", Model: "test", APIKey: "opaque-credential", MaxTokens: 512, Status: true}
	if err := db.Create(&config).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&aiModel.WorkflowConfig{ID: 1, Enabled: true, ModelID: config.ID, TimeoutSeconds: 10}).Error; err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("claims", &systemReq.CustomClaims{BaseClaims: systemReq.BaseClaims{ID: 7}})
	})
	router.POST("/autoCode/aiWorkflowChat", new(AutoCodeApi).WorkflowChat)
	callMode := func(chatMode, mode string) *httptest.ResponseRecorder {
		t.Helper()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/autoCode/aiWorkflowChat", strings.NewReader(`{"mode":"`+chatMode+`","query":"分析分类 password=hidden","response_mode":"`+mode+`","history":[{"role":"user","content":"上次需求"}]}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(rec, req)
		return rec
	}
	call := func(mode string) *httptest.ResponseRecorder { return callMode("analysisChat", mode) }
	blocking := call("blocking")
	var body struct {
		Code int `json:"code"`
		Data struct {
			Answer string `json:"answer"`
		} `json:"data"`
	}
	if err := json.Unmarshal(blocking.Body.Bytes(), &body); err != nil || body.Code != 0 || !strings.Contains(body.Data.Answer, "分类需求") {
		t.Fatalf("blocking failed: %s", blocking.Body.String())
	}
	if strings.Contains(body.Data.Answer, config.APIKey) || strings.Contains(submitted, "hidden") || !strings.Contains(submitted, "上次需求") {
		t.Fatal("context or redaction failure")
	}
	streaming := call("streaming")
	if !strings.Contains(streaming.Header().Get("Content-Type"), "text/event-stream") {
		t.Fatalf("not streaming: %s", streaming.Body.String())
	}
	var deltas strings.Builder
	var done bool
	for _, line := range strings.Split(streaming.Body.String(), "\n") {
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(line, "data:"))), &event); err != nil {
			t.Fatal(err)
		}
		if delta, ok := event["delta"].(string); ok {
			deltas.WriteString(delta)
		}
		if event["event"] == "done" {
			done = true
		}
	}
	if !done || strings.Contains(deltas.String(), config.APIKey) {
		t.Fatalf("bad stream or split-key leak: %s", deltas.String())
	}
	planning := callMode("workflowPromptChat", "blocking")
	if err := json.Unmarshal(planning.Body.Bytes(), &body); err != nil || body.Code != 0 || !strings.Contains(body.Data.Answer, "检查字段") || strings.Contains(body.Data.Answer, `"autoExecutable": true`) {
		t.Fatalf("planning failed: %s", planning.Body.String())
	}
	invalid = true
	failed := call("streaming")
	if !strings.Contains(failed.Body.String(), `"event":"error"`) || strings.Contains(failed.Body.String(), `"event":"done"`) {
		t.Fatalf("invalid result succeeded: %s", failed.Body.String())
	}
}
