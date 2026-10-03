package blog

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/example"
	aiService "github.com/isgvto/gin-vue-admin-gblog/server/service/ai"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestVisualLeaseMySQLNull(t *testing.T) {
	// Inspect the real MySQL dialect without opening a network connection.
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: "test:test@tcp(127.0.0.1:1)/test", SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true, DryRun: true, SkipDefaultTransaction: true})
	if err != nil {
		t.Fatal(err)
	}
	conn, _ := db.DB()
	defer conn.Close()
	task := blogModel.AiVisualTask{ID: strings.Repeat("a", 48), UserID: 1, Kind: "cover", CreatedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}
	insert := db.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx.Create(&task) })
	if !strings.Contains(insert, "`uploading_until`") || !strings.Contains(insert, "NULL") || strings.Contains(insert, "0000-00-00") {
		t.Fatalf("new preview inserts zero date: %s", insert)
	}
	release := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&blogModel.AiVisualTask{}).Where("id = ?", task.ID).Update("uploading_until", nil)
	})
	if !strings.Contains(release, "`uploading_until`=NULL") || strings.Contains(release, "0000-00-00") {
		t.Fatalf("release inserts zero date: %s", release)
	}
}

func visualTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	oldDB, oldCfg, oldLog := global.GVA_DB, global.GVA_CONFIG, global.GVA_LOG
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	conn, _ := db.DB()
	conn.SetMaxOpenConns(1)
	if err = db.AutoMigrate(&aiModel.AiModelConfig{}, &aiModel.ImageConfig{}, &blogModel.AiVisualTask{}, &example.ExaFileUploadAndDownload{}); err != nil {
		t.Fatal(err)
	}
	global.GVA_DB = db
	global.GVA_LOG = zap.NewNop()
	aiService.Factory().Invalidate()
	t.Cleanup(func() {
		global.GVA_DB = oldDB
		global.GVA_CONFIG = oldCfg
		global.GVA_LOG = oldLog
		aiService.Factory().Invalidate()
		conn.Close()
	})
	return db
}
func testVisualPNG(t *testing.T) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 12, 12))); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func TestVisualGeneratePreviewAndAdoptS3(t *testing.T) {
	db := visualTestDB(t)
	pngData := testVisualPNG(t)
	var generated, uploaded atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "images/generations") {
			generated.Add(1)
			if r.Header.Get("Authorization") != "Bearer image-key" {
				t.Error("wrong image credentials")
			}
			var req map[string]any
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req["model"] != "gpt-image-test" {
				t.Errorf("wrong image model: %v", req)
			}
			if _, exists := req["response_format"]; exists {
				t.Error("GPT image must not receive response_format")
			}
			json.NewEncoder(w).Encode(map[string]any{"data": []any{map[string]string{"b64_json": base64.StdEncoding.EncodeToString(pngData)}}})
			return
		}
		var req struct {
			Messages []struct{ Role, Content string }
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		var input VisualRequest
		_ = json.Unmarshal([]byte(req.Messages[len(req.Messages)-1].Content), &input)
		kind := input.Kind
		if kind == "auto" {
			kind = "flowchart"
		}
		plan, _ := json.Marshal(VisualPlan{Kind: kind, Prompt: "技术博客的简洁配图", Mermaid: "flowchart TD\n A[\"开始\"] --> B[\"完成\"]", Reason: "内容包含步骤"})
		json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"index": 0, "message": map[string]any{"role": "assistant", "content": string(plan)}, "finish_reason": "stop"}}})
	}))
	defer provider.Close()
	s3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			uploaded.Add(1)
			if r.Header.Get("Content-Type") != "image/png" {
				t.Error("missing image MIME")
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer s3.Close()
	db.Create(&aiModel.AiModelConfig{Name: "text", Provider: "openai", BaseURL: provider.URL, APIKey: "text-key", Model: "text-model", Status: true, IsDefault: true, MaxTokens: 4000})
	imageModel := aiModel.AiModelConfig{Name: "image", Provider: "openai", BaseURL: provider.URL, APIKey: "image-key", Model: "gpt-image-test", Status: true, MaxTokens: 4000}
	db.Create(&imageModel)
	db.Create(&aiModel.ImageConfig{ID: 1, Enabled: true, Provider: imageModel.Provider, BaseURL: imageModel.BaseURL, Model: imageModel.Model, APIKey: imageModel.APIKey, TimeoutSeconds: 30})
	global.GVA_CONFIG.System.OssType = "aws-s3"
	global.GVA_CONFIG.AwsS3.Endpoint = s3.URL
	global.GVA_CONFIG.AwsS3.Bucket = "test"
	global.GVA_CONFIG.AwsS3.Region = "us-east-1"
	global.GVA_CONFIG.AwsS3.SecretID = "test"
	global.GVA_CONFIG.AwsS3.SecretKey = "test"
	global.GVA_CONFIG.AwsS3.S3ForcePathStyle = true
	global.GVA_CONFIG.AwsS3.BaseURL = "https://images.example.test"
	global.GVA_CONFIG.AwsS3.PathPrefix = "blog"
	svc := &VisualService{}
	req := VisualRequest{Kind: "auto", Title: "测试博客", Source: "开始后完成", Size: "1024x1024"}
	plan, err := svc.Plan(context.Background(), req)
	if err != nil || plan.Kind != "flowchart" {
		t.Fatalf("plan: %+v %v", plan, err)
	}
	req.Kind = "cover"
	result, err := svc.Generate(context.Background(), 7, req)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(result.Preview, "data:image/png;base64,") || uploaded.Load() != 0 || generated.Load() != 1 {
		t.Fatal("preview must not upload")
	}
	var nullLeases int64
	db.Model(&blogModel.AiVisualTask{}).Where("id = ? AND uploading_until IS NULL", result.ID).Count(&nullLeases)
	if nullLeases != 1 {
		t.Fatal("new preview lease must be SQL NULL")
	}
	var files int64
	db.Model(&example.ExaFileUploadAndDownload{}).Count(&files)
	if files != 0 {
		t.Fatal("preview created attachment")
	}
	if _, err = svc.Adopt(context.Background(), 8, VisualAdoptRequest{ID: result.ID}); err == nil {
		t.Fatal("foreign preview accepted")
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = svc.Adopt(cancelled, 7, VisualAdoptRequest{ID: result.ID}); err == nil {
		t.Fatal("cancelled upload succeeded")
	}
	var retained blogModel.AiVisualTask
	db.First(&retained, "id = ?", result.ID)
	if len(retained.PNG) == 0 || retained.URL != "" || retained.UploadingUntil != nil {
		t.Fatal("failed upload lost preview or retained lease")
	}
	adopted, err := svc.Adopt(context.Background(), 7, VisualAdoptRequest{ID: result.ID})
	if err != nil {
		t.Fatal(err)
	}
	again, err := svc.Adopt(context.Background(), 7, VisualAdoptRequest{ID: result.ID})
	if err != nil || again.URL != adopted.URL || uploaded.Load() != 1 || generated.Load() != 1 {
		t.Fatalf("adopt not idempotent: %v", err)
	}
	db.Model(&example.ExaFileUploadAndDownload{}).Count(&files)
	if files != 1 {
		t.Fatalf("attachments: %d", files)
	}
	var saved blogModel.AiVisualTask
	db.First(&saved, "id = ?", result.ID)
	if len(saved.PNG) != 0 || saved.FileID == 0 || saved.UploadingUntil != nil {
		t.Fatal("temporary bytes not released")
	}
	req.Kind = "flowchart"
	diagram, err := svc.Generate(context.Background(), 7, req)
	if err != nil || diagram.Mermaid == "" || diagram.Preview != "" || generated.Load() != 1 {
		t.Fatalf("diagram used image endpoint: %v", err)
	}
	if _, err = svc.Adopt(context.Background(), 7, VisualAdoptRequest{ID: diagram.ID, PNG: base64.StdEncoding.EncodeToString(pngData)}); err != nil {
		t.Fatal(err)
	}
}

func TestImageSupplierFailureFormats(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		body   string
	}{
		{"denied", 401, `{"error":{"message":"secret-key"}}`},
		{"url-only", 200, `{"data":[{"url":"http://127.0.0.1/private"}]}`},
		{"invalid-base64", 200, `{"data":[{"b64_json":"invalid?"}]}`},
		{"not-image", 200, `{"data":[{"b64_json":"PHN2Zz48L3N2Zz4="}]}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			}))
			defer endpoint.Close()
			cfg := aiService.ImageEndpointConfig{BaseURL: endpoint.URL, Model: "gpt-image-test", APIKey: "secret-key", TimeoutSeconds: 10}
			_, err := (&aiService.ImageService{}).GenerateWithConfig(context.Background(), cfg, "测试配图", "1024x1024")
			if err == nil || strings.Contains(err.Error(), "secret-key") {
				t.Fatalf("unsafe supplier failure: %v", err)
			}
		})
	}
}
func TestVisualExpiryValidationAndRetry(t *testing.T) {
	db := visualTestDB(t)
	svc := &VisualService{}
	id := strings.Repeat("a", 48)
	task := blogModel.AiVisualTask{ID: id, UserID: 7, Kind: "cover", PNG: testVisualPNG(t), ExpiresAt: time.Now().Add(-time.Minute)}
	db.Create(&task)
	if _, err := svc.Adopt(context.Background(), 7, VisualAdoptRequest{ID: id}); err == nil {
		t.Fatal("expired preview accepted")
	}
	if err := CleanupVisualTasks(db); err != nil {
		t.Fatal(err)
	}
	var count int64
	db.Model(&blogModel.AiVisualTask{}).Count(&count)
	if count != 0 {
		t.Fatal("expired preview not cleaned")
	}
	for _, req := range []VisualRequest{{Kind: "unknown", Source: "x", Size: "1024x1024"}, {Kind: "cover", Size: "1024x1024"}, {Kind: "flowchart", Source: "x", Size: "bad"}} {
		if ValidateVisualRequest(req) == nil {
			t.Fatal("invalid request accepted")
		}
	}
	for _, source := range []string{"flowchart TD\n%%{init: {}}%%\nA-->B", "flowchart TD\n%%   {init: {}}%%\nA-->B", "flowchart TD\nA[<script>]", "sequenceDiagram\nA->>B: x"} {
		if ValidateVisualPlan(VisualPlan{Kind: "flowchart", Prompt: "x", Mermaid: source}, "flowchart") == nil {
			t.Fatal("unsafe diagram accepted")
		}
	}
	if _, err := aiService.NormalizeImage([]byte("<svg onload='bad()'></svg>")); err == nil {
		t.Fatal("SVG upload accepted")
	}
	model := aiModel.AiModelConfig{Name: "text", Model: "text-model", Provider: "openai", APIKey: "text-key", BaseURL: "https://text.example.test", Status: true, IsDefault: true, MaxTokens: 4000}
	db.Create(&model)
	service := &aiService.ImageService{}
	if err := service.Save(aiService.ImageConfigInput{Enabled: true, Provider: "ark", Model: "picture", APIKey: "image-secret", TimeoutSeconds: 30}); err != nil {
		t.Fatal(err)
	}
	loaded, _ := service.Config()
	endpoint, err := service.Resolve(loaded)
	if err != nil || endpoint.Model != "picture" || endpoint.APIKey != "image-secret" {
		t.Fatalf("independent image resolve: %v", err)
	}
	db.Model(&model).Updates(map[string]any{"model": "changed", "status": false})
	endpoint, err = service.Resolve(loaded)
	if err != nil || endpoint.Model != "picture" {
		t.Fatal("shared model change affected images")
	}
	if err = service.Save(aiService.ImageConfigInput{Enabled: true, Provider: "openai", Model: "picture", TimeoutSeconds: 30}); err == nil {
		t.Fatal("empty URL accepted")
	}
	serialized, _ := json.Marshal(loaded)
	if bytes.Contains(serialized, []byte("secret")) {
		t.Fatal("key leaked")
	}
}
