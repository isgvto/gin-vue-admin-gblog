package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	aiReq "github.com/isgvto/gin-vue-admin-gblog/server/model/ai/request"
)

const probeTimeout = 30 * time.Second

// resolveProbe uses the submitted draft; only an omitted key is taken from storage.
func (s *ModelConfigService) resolveProbe(info aiReq.AiModelConfigUpsert) (*aiModel.AiModelConfig, error) {
	info.BaseURL = strings.TrimRight(strings.TrimSpace(info.BaseURL), "/")
	info.Model = strings.TrimSpace(info.Model)
	if info.ID != 0 {
		var saved aiModel.AiModelConfig
		if err := global.GVA_DB.First(&saved, info.ID).Error; err != nil {
			return nil, errors.New("配置不存在")
		}
		if info.APIKey == "" {
			info.APIKey = saved.APIKey
		}
	}
	if info.Provider != "openai" && info.Provider != "ark" && info.Provider != "gemini" {
		return nil, errors.New("不支持的供应商")
	}
	if strings.TrimSpace(info.APIKey) == "" {
		return nil, errors.New("请填写 API Key")
	}
	if info.Provider == "openai" && info.BaseURL == "" {
		return nil, errors.New("请填写 Base URL")
	}
	return &aiModel.AiModelConfig{Provider: info.Provider, BaseURL: info.BaseURL, APIKey: info.APIKey, Model: info.Model, Temperature: info.Temperature, MaxTokens: info.MaxTokens}, nil
}

func safeProbeError(err error, key string) error {
	if err == nil {
		return nil
	}
	message := err.Error()
	if key != "" {
		message = strings.ReplaceAll(message, key, "[隐藏密钥]")
		message = strings.ReplaceAll(message, url.QueryEscape(key), "[隐藏密钥]")
	}
	return errors.New(message)
}

// probeConfig performs a real generation using the same adapter and parameters as chat.
func probeConfig(ctx context.Context, cfg *aiModel.AiModelConfig) error {
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	cm, err := Factory().build(ctx, cfg)
	if err == nil {
		var reply *schema.Message
		reply, err = cm.Generate(ctx, []*schema.Message{schema.UserMessage("Reply with OK only.")})
		if err == nil && (reply == nil || (strings.TrimSpace(reply.Content) == "" && strings.TrimSpace(reply.ReasoningContent) == "")) {
			err = errors.New("供应商返回空响应，模型未生成有效内容")
		}
	}
	return safeProbeError(err, cfg.APIKey)
}

func (s *ModelConfigService) TestConnection(ctx context.Context, info aiReq.AiModelConfigUpsert) error {
	cfg, err := s.resolveProbe(info)
	if err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return errors.New("请选择或填写模型标识")
	}
	if info.MaxTokens <= 0 || info.Temperature < 0 || info.Temperature > 2 {
		return errors.New("请填写有效的温度和最大输出参数")
	}
	return probeConfig(ctx, cfg)
}

func (f *ModelFactory) CheckConnection(ctx context.Context) (bool, string) {
	f.Invalidate()
	cfg, err := f.ActiveConfig()
	if err != nil {
		return false, err.Error()
	}
	if err := probeConfig(ctx, cfg); err != nil {
		return false, "模型连接测试失败：" + err.Error()
	}
	return true, cfg.Name
}

type ProviderModel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *ModelConfigService) ListProviderModels(ctx context.Context, info aiReq.AiModelConfigUpsert) ([]ProviderModel, error) {
	cfg, err := s.resolveProbe(info)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	base := strings.TrimRight(cfg.BaseURL, "/")
	if cfg.Provider == "ark" && base == "" {
		base = "https://ark.cn-beijing.volces.com/api/v3"
	}
	if cfg.Provider == "gemini" {
		base = "https://generativelanguage.googleapis.com/v1beta"
	}
	endpoint, err := url.Parse(base + "/models")
	if err != nil || (endpoint.Scheme != "http" && endpoint.Scheme != "https") || endpoint.Host == "" {
		return nil, errors.New("Base URL 必须是有效的 HTTP 或 HTTPS 地址")
	}
	client := &http.Client{Timeout: probeTimeout, CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}
	models := make([]ProviderModel, 0)
	seen := map[string]bool{}
	pageToken := ""
	for page := 0; page < 100; page++ {
		query := endpoint.Query()
		if cfg.Provider == "gemini" {
			query.Set("pageSize", "1000")
			query.Set("pageToken", pageToken)
		}
		endpoint.RawQuery = query.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
		if err != nil {
			return nil, safeProbeError(err, cfg.APIKey)
		}
		if cfg.Provider == "gemini" {
			req.Header.Set("x-goog-api-key", cfg.APIKey)
		} else {
			req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, safeProbeError(err, cfg.APIKey)
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, (2<<20)+1))
		resp.Body.Close()
		if readErr != nil {
			return nil, safeProbeError(readErr, cfg.APIKey)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if resp.StatusCode == 404 || resp.StatusCode == 405 {
				return nil, errors.New("供应商未提供兼容的模型列表接口，请手动填写模型或接入点 ID")
			}
			return nil, fmt.Errorf("读取模型列表失败（HTTP %d），请检查密钥、地址和供应商权限", resp.StatusCode)
		}
		if len(body) > 2<<20 {
			return nil, errors.New("模型列表响应过大")
		}
		var result struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
			Models []struct {
				Name        string   `json:"name"`
				DisplayName string   `json:"displayName"`
				Methods     []string `json:"supportedGenerationMethods"`
			} `json:"models"`
			NextPageToken string `json:"nextPageToken"`
		}
		if json.Unmarshal(body, &result) != nil {
			return nil, errors.New("供应商返回的模型列表格式无效")
		}
		add := func(id, name string) {
			if id != "" && !seen[id] {
				if name == "" {
					name = id
				}
				seen[id] = true
				models = append(models, ProviderModel{ID: id, Name: name})
			}
		}
		if cfg.Provider == "gemini" {
			for _, m := range result.Models {
				for _, method := range m.Methods {
					if method == "generateContent" {
						add(strings.TrimPrefix(m.Name, "models/"), m.DisplayName)
						break
					}
				}
			}
			if result.NextPageToken != "" {
				if result.NextPageToken == pageToken {
					return nil, errors.New("供应商模型列表分页异常")
				}
				pageToken = result.NextPageToken
				continue
			}
		} else {
			for _, m := range result.Data {
				add(m.ID, m.ID)
			}
		}
		sort.Slice(models, func(i, j int) bool { return models[i].ID < models[j].ID })
		return models, nil
	}
	return nil, errors.New("供应商模型列表分页过多")
}
