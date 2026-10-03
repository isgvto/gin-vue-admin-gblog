package ai

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/gemini"
	"github.com/cloudwego/eino-ext/components/model/openai"
	fmodel "github.com/cloudwego/eino/components/model"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	"go.uber.org/zap"
	"google.golang.org/genai"
)

// ModelFactory 按数据库中的默认模型配置动态创建并缓存 ChatModel。
// 缓存键包含模型参数与密钥的摘要，配置变更后自然失效，无需重启。
type ModelFactory struct {
	mu       sync.RWMutex
	cached   fmodel.ToolCallingChatModel
	cacheKey string

	ttlMu     sync.Mutex
	lastFetch time.Time
	lastCfg   *aiModel.AiModelConfig
}

const activeConfigTTL = 60 * time.Second

// ActiveConfig 返回当前生效的模型配置：优先取数据库默认启用配置，否则回退 config.yaml 兜底段。
func (f *ModelFactory) ActiveConfig() (*aiModel.AiModelConfig, error) {
	f.ttlMu.Lock()
	defer f.ttlMu.Unlock()

	if f.lastCfg != nil && time.Since(f.lastFetch) < activeConfigTTL {
		return f.lastCfg, nil
	}

	var err error
	var cfg *aiModel.AiModelConfig
	if global.GVA_DB != nil {
		var dbCfg aiModel.AiModelConfig
		if e := global.GVA_DB.Where("is_default = ? AND status = ?", true, true).First(&dbCfg).Error; e == nil {
			cfg = &dbCfg
		}
	}
	if cfg == nil && global.GVA_CONFIG.AI.Enable && global.GVA_CONFIG.AI.APIKey != "" {
		y := global.GVA_CONFIG.AI
		cfg = &aiModel.AiModelConfig{
			Name: "yaml-" + y.Model, Provider: y.Provider, BaseURL: y.BaseURL,
			APIKey: y.APIKey, Model: y.Model, Temperature: y.Temperature, MaxTokens: y.MaxTokens, Status: true,
		}
	}
	if cfg == nil {
		err = fmt.Errorf("未配置可用的大模型，请先在 AI 模型配置中设置默认模型")
	} else {
		f.lastCfg = cfg
		f.lastFetch = time.Now()
	}
	return cfg, err
}

// Get 返回当前生效配置对应的 ChatModel 实例（带缓存）。
func (f *ModelFactory) Get(ctx context.Context) (fmodel.ToolCallingChatModel, error) {
	cfg, err := f.ActiveConfig()
	if err != nil {
		return nil, err
	}

	key := modelConfigCacheKey(cfg)
	f.mu.RLock()
	if f.cacheKey == key && f.cached != nil {
		defer f.mu.RUnlock()
		return f.cached, nil
	}
	f.mu.RUnlock()

	cm, err := f.build(ctx, cfg)
	if err != nil {
		return nil, err
	}

	f.mu.Lock()
	f.cached = cm
	f.cacheKey = key
	f.mu.Unlock()
	return cm, nil
}

func modelConfigCacheKey(cfg *aiModel.AiModelConfig) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d|%q|%q|%q|%q|%g|%d", cfg.ID, cfg.Provider, cfg.BaseURL, cfg.APIKey, cfg.Model, cfg.Temperature, cfg.MaxTokens))))
}

func (f *ModelFactory) build(ctx context.Context, cfg *aiModel.AiModelConfig) (fmodel.ToolCallingChatModel, error) {
	temp := cfg.Temperature
	maxTokens := cfg.MaxTokens
	switch cfg.Provider {
	case "openai":
		return openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey:      cfg.APIKey,
			BaseURL:     cfg.BaseURL,
			Model:       cfg.Model,
			Temperature: &temp,
			MaxTokens:   &maxTokens,
		})
	case "ark":
		return ark.NewChatModel(ctx, &ark.ChatModelConfig{
			APIKey:      cfg.APIKey,
			BaseURL:     cfg.BaseURL,
			Model:       cfg.Model,
			Temperature: &temp,
			MaxTokens:   &maxTokens,
		})
	case "gemini":
		client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.APIKey})
		if err != nil {
			return nil, fmt.Errorf("创建 Gemini 客户端失败: %w", err)
		}
		return gemini.NewChatModel(ctx, &gemini.Config{
			Client:      client,
			Model:       cfg.Model,
			Temperature: &temp,
			MaxTokens:   &maxTokens,
		})
	default:
		return nil, fmt.Errorf("不支持的模型供应商: %s", cfg.Provider)
	}
}

// Available 用于探活：默认模型是否存在、能否构建实例。
func (f *ModelFactory) Available() (bool, string) {
	cfg, err := f.ActiveConfig()
	if err != nil {
		return false, err.Error()
	}
	return true, cfg.Name
}

// Invalidate 在模型配置变更后调用，立即清空配置缓存，保证配置页操作即时生效。
func (f *ModelFactory) Invalidate() {
	f.ttlMu.Lock()
	f.lastCfg = nil
	f.lastFetch = time.Time{}
	f.ttlMu.Unlock()
	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("AI 模型配置缓存已失效", zap.String("op", "invalidate"))
	}
}
