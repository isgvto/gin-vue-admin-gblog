package ai

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	"gorm.io/gorm"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndependentImageAPIKeepsSecretAndTestDoesNotSave(t *testing.T) {
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
	db.Create(&aiModel.ImageConfig{ID: 1, Provider: "ark", Model: "saved-image", APIKey: "private-key", TimeoutSeconds: 180})
	var buffer bytes.Buffer
	_ = png.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, 8, 8)))
	calls := 0
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Bearer private-key" {
			t.Error("saved image key not used")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{map[string]string{"b64_json": base64.StdEncoding.EncodeToString(buffer.Bytes())}}})
	}))
	defer provider.Close()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := &ModelConfigApi{}
	router.GET("/config", api.GetImageConfig)
	router.POST("/test", api.TestImageConfig)
	get := httptest.NewRecorder()
	router.ServeHTTP(get, httptest.NewRequest("GET", "/config", nil))
	if strings.Contains(get.Body.String(), "private-key") || !strings.Contains(get.Body.String(), `"hasKey":true`) || strings.Contains(get.Body.String(), "modelId") {
		t.Fatalf("unsafe config response: %s", get.Body.String())
	}
	raw, _ := json.Marshal(map[string]any{"provider": "ark", "baseUrl": provider.URL, "model": "draft-image", "apiKey": "", "timeoutSeconds": 10})
	res := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(res, req)
	if calls != 1 || !strings.Contains(res.Body.String(), `"code":0`) {
		t.Fatalf("image test failed: %s", res.Body.String())
	}
	var cfg aiModel.ImageConfig
	db.First(&cfg, 1)
	if cfg.Model != "saved-image" || cfg.APIKey != "private-key" || cfg.Enabled {
		t.Fatal("test saved draft")
	}
}
