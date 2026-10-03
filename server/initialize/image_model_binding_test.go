package initialize

import (
	"github.com/glebarez/sqlite"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	"gorm.io/gorm"
	"testing"
)

func TestImageIndependentMigration(t *testing.T) {
	for _, mode := range []string{"binding", "default", "existing", "missing"} {
		t.Run(mode, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			if err != nil {
				t.Fatal(err)
			}
			if err = db.AutoMigrate(&aiModel.ImageConfig{}, &aiModel.AiModelConfig{}); err != nil {
				t.Fatal(err)
			}
			model := aiModel.AiModelConfig{Name: "original", Provider: "ark", Model: "picture", APIKey: "secret", Status: true, IsDefault: true, MaxTokens: 4000}
			db.Create(&model)
			cfg := aiModel.ImageConfig{ID: 1, Enabled: true, ModelID: model.ID, TimeoutSeconds: 180}
			if mode == "default" {
				cfg.ModelID = 0
			}
			if mode == "existing" {
				cfg.Model = "own"
				cfg.APIKey = "own-key"
				cfg.Provider = "openai"
				cfg.BaseURL = "https://own.example.test"
			}
			if mode == "missing" {
				cfg.ModelID = 999
			}
			db.Create(&cfg)
			if err = migrateImageModelBinding(db); err != nil {
				t.Fatal(err)
			}
			db.First(&cfg, 1)
			if !cfg.IndependentMigrated || cfg.ModelID != 0 {
				t.Fatal("binding not retired")
			}
			if mode == "missing" {
				if cfg.Enabled {
					t.Fatal("missing binding remained enabled")
				}
			} else if mode == "existing" {
				if cfg.Model != "own" || cfg.APIKey != "own-key" {
					t.Fatal("existing credentials overwritten")
				}
			} else if cfg.Model != "picture" || cfg.APIKey != "secret" {
				t.Fatal("assignment not copied")
			}
			db.Model(&model).Update("model", "changed")
			if err = migrateImageModelBinding(db); err != nil {
				t.Fatal(err)
			}
			var after aiModel.ImageConfig
			db.First(&after, 1)
			if after.Model != cfg.Model {
				t.Fatal("migration repeated")
			}
			var count int64
			db.Model(&aiModel.AiModelConfig{}).Count(&count)
			if count != 1 {
				t.Fatal("shared records altered")
			}
		})
	}
}
