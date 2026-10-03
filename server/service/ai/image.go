package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	_ "golang.org/x/image/webp"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ImageService struct{}
type ImageConfigInput struct {
	Enabled        bool   `json:"enabled"`
	Provider       string `json:"provider"`
	BaseURL        string `json:"baseUrl"`
	Model          string `json:"model"`
	APIKey         string `json:"apiKey"`
	ClearKey       bool   `json:"clearKey"`
	TimeoutSeconds int    `json:"timeoutSeconds"`
}

// ImageEndpointConfig is resolved from the independent image configuration.
// It is never persisted or returned by configuration APIs.
type ImageEndpointConfig struct {
	Provider       string
	BaseURL        string
	Model          string
	APIKey         string
	TimeoutSeconds int
}

const defaultArkBaseURL = "https://ark.cn-beijing.volces.com/api/v3"

func resolveImageEndpoint(model *aiModel.AiModelConfig, timeout int) (ImageEndpointConfig, error) {
	base := strings.TrimRight(strings.TrimSpace(model.BaseURL), "/")
	if base == "" && model.Provider == "ark" {
		base = defaultArkBaseURL
	}
	endpoint := ImageEndpointConfig{Provider: model.Provider, BaseURL: base, Model: model.Model, APIKey: model.APIKey, TimeoutSeconds: timeout}
	return endpoint, ValidateImageConfig(endpoint)
}

func (*ImageService) Config() (aiModel.ImageConfig, error) {
	cfg := aiModel.ImageConfig{ID: 1, TimeoutSeconds: 180}
	err := global.GVA_DB.First(&cfg, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cfg, nil
	}
	return cfg, err
}
func ValidateImageConfig(cfg ImageEndpointConfig) error {
	if cfg.TimeoutSeconds < 10 || cfg.TimeoutSeconds > 600 {
		return errors.New("图片生成超时必须在10到600秒之间")
	}
	endpoint, err := url.Parse(strings.TrimSpace(cfg.BaseURL))
	if err != nil || endpoint.Host == "" || (endpoint.Scheme != "https" && endpoint.Scheme != "http") || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return errors.New("图片接口地址无效，请在独立图片配置中填写有效的 BaseURL；Ark 地址留空时使用默认地址")
	}
	if strings.TrimSpace(cfg.Model) == "" || strings.TrimSpace(cfg.APIKey) == "" {
		return errors.New("图片模型标识和 API Key 不能为空")
	}
	return nil
}
func (s *ImageService) Save(input ImageConfigInput) error {
	cfg, err := s.Config()
	if err != nil {
		return err
	}
	cfg.Enabled = input.Enabled
	cfg.ModelID = 0
	cfg.IndependentMigrated = true
	cfg.Provider = strings.TrimSpace(input.Provider)
	if cfg.Provider == "" {
		cfg.Provider = "openai"
	}
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(input.BaseURL), "/")
	cfg.Model = strings.TrimSpace(input.Model)
	if input.ClearKey {
		cfg.APIKey = ""
	} else if strings.TrimSpace(input.APIKey) != "" {
		cfg.APIKey = strings.TrimSpace(input.APIKey)
	}
	if cfg.Provider != "openai" && cfg.Provider != "ark" {
		return errors.New("请选择 OpenAI 兼容或 Ark 图片接口")
	}
	cfg.TimeoutSeconds = input.TimeoutSeconds
	if cfg.Enabled {
		if _, err := s.Resolve(cfg); err != nil {
			return err
		}
	} else if cfg.TimeoutSeconds < 10 || cfg.TimeoutSeconds > 600 {
		return errors.New("超时必须在10到600秒之间")
	}
	return global.GVA_DB.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.AssignmentColumns([]string{"enabled", "model_id", "provider", "base_url", "model", "api_key", "timeout_seconds", "independent_migrated"})}).Create(&cfg).Error
}

func (*ImageService) Resolve(cfg aiModel.ImageConfig) (ImageEndpointConfig, error) {
	return resolveImageEndpoint(&aiModel.AiModelConfig{Provider: cfg.Provider, BaseURL: cfg.BaseURL, Model: cfg.Model, APIKey: cfg.APIKey}, cfg.TimeoutSeconds)
}

// NormalizeImage prevents HTML/SVG payloads and excessive decoded pixel sizes.
func NormalizeImage(data []byte) ([]byte, error) {
	if len(data) == 0 || len(data) > 10<<20 {
		return nil, errors.New("图片为空或超过10MB")
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "png" && format != "jpeg" && format != "webp") || cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > 4096 || cfg.Height > 4096 || cfg.Width*cfg.Height > 16*1024*1024 {
		return nil, errors.New("图片格式或尺寸无效，仅支持 PNG、JPEG、WebP")
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, errors.New("无法解码图片")
	}
	var out bytes.Buffer
	if err = png.Encode(&out, img); err != nil {
		return nil, err
	}
	if out.Len() > 10<<20 {
		return nil, errors.New("图片超过10MB")
	}
	return out.Bytes(), nil
}
func (s *ImageService) Generate(ctx context.Context, prompt, size string) ([]byte, error) {
	cfg, err := s.Config()
	if err != nil {
		return nil, err
	}
	if !cfg.Enabled {
		return nil, errors.New("图片生成未启用，请在 AI 模型配置中配置图片模型")
	}
	endpoint, err := s.Resolve(cfg)
	if err != nil {
		return nil, err
	}
	return s.GenerateWithConfig(ctx, endpoint, prompt, size)
}
func (*ImageService) GenerateWithConfig(ctx context.Context, cfg ImageEndpointConfig, prompt, size string) ([]byte, error) {
	if err := ValidateImageConfig(cfg); err != nil {
		return nil, err
	}
	if size != "1024x1024" && size != "1536x1024" && size != "1024x1536" {
		return nil, errors.New("不支持的图片尺寸")
	}
	if len([]rune(prompt)) == 0 || len([]rune(prompt)) > 6000 {
		return nil, errors.New("配图描述应在1到6000字之间")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.TimeoutSeconds)*time.Second)
	defer cancel()
	payload, err := imageRequestPayload(cfg, prompt, size)
	if err != nil {
		return nil, err
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(cfg.BaseURL, "/")+"/images/generations", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	res, err := client.Do(req)
	if err != nil {
		return nil, analysisError(err, cfg.APIKey)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, imageProviderError(res, cfg, fmt.Sprint(payload["size"]))
	}
	raw, err := io.ReadAll(io.LimitReader(res.Body, 20<<20+1))
	if err != nil {
		return nil, errors.New("读取图片响应失败")
	}
	if len(raw) > 20<<20 {
		return nil, errors.New("图片响应过大")
	}
	var result struct {
		Data []struct {
			B64 string `json:"b64_json"`
		} `json:"data"`
	}
	if json.Unmarshal(raw, &result) != nil || len(result.Data) == 0 || result.Data[0].B64 == "" {
		return nil, errors.New("供应商未返回 Base64 图片，请使用支持 OpenAI 图片生成格式的接口")
	}
	data, err := base64.StdEncoding.DecodeString(result.Data[0].B64)
	if err != nil {
		return nil, errors.New("供应商图片编码无效")
	}
	return NormalizeImage(data)
}

// Return bounded, structured diagnostics rather than discarding the supplier's
// reason or exposing arbitrary HTML, response headers, prompts and credentials.
func imageProviderError(res *http.Response, cfg ImageEndpointConfig, size string) error {
	message := fmt.Sprintf("图片供应商返回 HTTP %d", res.StatusCode)
	if res.StatusCode == http.StatusTooManyRequests {
		message = "图片供应商限流（HTTP 429），请稍后重试或检查供应商配额"
	}
	if res.StatusCode == http.StatusNotFound {
		message = "图片接口不存在（HTTP 404），请检查 BaseURL 与 /images/generations 接口"
	}
	raw, err := io.ReadAll(io.LimitReader(res.Body, (16<<10)+1))
	if err == nil && len(raw) <= 16<<10 {
		var body struct {
			Message string          `json:"message"`
			Code    json.RawMessage `json:"code"`
			Error   json.RawMessage `json:"error"`
		}
		if json.Unmarshal(raw, &body) == nil {
			var detail struct {
				Message string          `json:"message"`
				Code    json.RawMessage `json:"code"`
				Param   string          `json:"param"`
			}
			if json.Unmarshal(body.Error, &detail) != nil {
				_ = json.Unmarshal(body.Error, &detail.Message)
			}
			if detail.Message == "" {
				detail.Message = body.Message
			}
			if len(detail.Code) == 0 {
				detail.Code = body.Code
			}
			if detail.Message != "" {
				message += "：" + detail.Message
			}
			var code string
			if json.Unmarshal(detail.Code, &code) == nil && code != "" {
				message += "\n错误码：" + code
			}
			if detail.Param != "" {
				message += "\n参数：" + detail.Param
			}
		}
	}
	message += fmt.Sprintf("\n模型：%s，请求尺寸：%s", cfg.Model, size)
	if cfg.APIKey != "" {
		message = strings.ReplaceAll(message, cfg.APIKey, "[已脱敏]")
	}
	message = SanitizeAnalysisText(message)
	runes := []rune(message)
	if len(runes) > 1000 {
		message = string(runes[:1000]) + "…"
	}
	return errors.New(message)
}
