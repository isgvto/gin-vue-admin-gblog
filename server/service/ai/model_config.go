package ai

import (
	"errors"
	"fmt"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	aiReq "github.com/isgvto/gin-vue-admin-gblog/server/model/ai/request"
	"gorm.io/gorm"
)

type ModelConfigService struct{}

func (s *ModelConfigService) factoryOrDefault() *ModelFactory {
	return Factory()
}

type ModelConfigItem struct {
	aiModel.AiModelConfig
	ID      uint   `json:"id"`
	HasKey  bool   `json:"hasKey"`
	KeyTail string `json:"keyTail"`
}

// toSafeItem 将配置转为对外项：不回传 api key，仅返回是否已配置与尾4位。
func toSafeItem(cfg aiModel.AiModelConfig) ModelConfigItem {
	item := ModelConfigItem{AiModelConfig: cfg, ID: cfg.ID}
	if cfg.APIKey != "" {
		item.HasKey = true
		runes := []rune(cfg.APIKey)
		if len(runes) > 4 {
			item.KeyTail = string(runes[len(runes)-4:])
		} else {
			item.KeyTail = "****"
		}
	}
	item.APIKey = ""
	return item
}

func (s *ModelConfigService) GetList(info aiReq.AiModelConfigSearch) (list []ModelConfigItem, total int64, err error) {
	db := global.GVA_DB.Model(&aiModel.AiModelConfig{})
	if info.Name != "" {
		db = db.Where("name LIKE ?", "%"+info.Name+"%")
	}
	if info.Provider != "" {
		db = db.Where("provider = ?", info.Provider)
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	var records []aiModel.AiModelConfig
	if err = db.Order("is_default DESC, id ASC").Scopes(paginate(info.Page, info.PageSize)).Find(&records).Error; err != nil {
		return
	}
	list = make([]ModelConfigItem, 0, len(records))
	for _, r := range records {
		list = append(list, toSafeItem(r))
	}
	return
}

func paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 || pageSize > 100 {
			pageSize = 10
		}
		return db.Offset((page - 1) * pageSize).Limit(pageSize)
	}
}

func (s *ModelConfigService) validate(info *aiReq.AiModelConfigUpsert) error {
	if info.Name == "" || info.Provider == "" || info.Model == "" {
		return errors.New("名称、供应商与模型标识不能为空")
	}
	if info.Provider != "openai" && info.Provider != "ark" && info.Provider != "gemini" {
		return fmt.Errorf("不支持的供应商: %s", info.Provider)
	}
	if info.Provider == "openai" && info.BaseURL == "" {
		return errors.New("OpenAI 兼容供应商必须填写 BaseURL")
	}
	if info.Temperature < 0 || info.Temperature > 2 {
		return errors.New("温度必须在 0 到 2 之间")
	}
	if info.MaxTokens <= 0 {
		return errors.New("最大输出 token 必须大于 0")
	}
	return nil
}

func (s *ModelConfigService) Create(info aiReq.AiModelConfigUpsert) error {
	if err := s.validate(&info); err != nil {
		return err
	}
	if info.APIKey == "" {
		return errors.New("API Key 不能为空")
	}
	record := aiModel.AiModelConfig{
		Name: info.Name, Provider: info.Provider, BaseURL: info.BaseURL,
		APIKey: info.APIKey, Model: info.Model,
		Temperature: info.Temperature, MaxTokens: info.MaxTokens,
		Status: info.Status, Remark: info.Remark,
	}
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if info.Status {
			// 首个启用配置自动成为默认，保证开箱可用
			var count int64
			if err := tx.Model(&aiModel.AiModelConfig{}).Where("is_default = ? AND status = ?", true, true).Count(&count).Error; err != nil {
				return err
			}
			record.IsDefault = count == 0
		}
		return tx.Create(&record).Error
	})
	if err == nil {
		s.factoryOrDefault().Invalidate()
	}
	return err
}

func (s *ModelConfigService) Update(info aiReq.AiModelConfigUpsert) error {
	if info.ID == 0 {
		return errors.New("缺少配置 ID")
	}
	if err := s.validate(&info); err != nil {
		return err
	}
	var record aiModel.AiModelConfig
	if err := global.GVA_DB.First(&record, info.ID).Error; err != nil {
		return errors.New("配置不存在")
	}
	if record.IsDefault && !info.Status {
		return errors.New("默认模型不可停用，请先将其他模型设为默认")
	}
	updates := map[string]any{
		"name": info.Name, "provider": info.Provider, "base_url": info.BaseURL,
		"model": info.Model, "temperature": info.Temperature,
		"max_tokens": info.MaxTokens, "status": info.Status, "remark": info.Remark,
	}
	if info.APIKey != "" {
		updates["api_key"] = info.APIKey
	}
	if err := global.GVA_DB.Model(&record).Updates(updates).Error; err != nil {
		return err
	}
	s.factoryOrDefault().Invalidate()
	return nil
}

func (s *ModelConfigService) Delete(id uint) error {
	if id == 0 {
		return errors.New("无效的配置 ID")
	}
	var record aiModel.AiModelConfig
	if err := global.GVA_DB.First(&record, id).Error; err != nil {
		return errors.New("配置不存在")
	}
	if record.IsDefault {
		return errors.New("默认模型不可删除，请先将其他模型设为默认")
	}
	if err := global.GVA_DB.Delete(&record).Error; err != nil {
		return err
	}
	s.factoryOrDefault().Invalidate()
	return nil
}

func (s *ModelConfigService) SetDefault(id uint) error {
	if id == 0 {
		return errors.New("无效的配置 ID")
	}
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var record aiModel.AiModelConfig
		if err := tx.First(&record, id).Error; err != nil {
			return errors.New("配置不存在")
		}
		if !record.Status {
			return errors.New("停用中的模型不可设为默认")
		}
		if err := tx.Model(&aiModel.AiModelConfig{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return err
		}
		if err := tx.Model(&record).Update("is_default", true).Error; err != nil {
			return err
		}
		return nil
	})
	if err == nil {
		s.factoryOrDefault().Invalidate()
	}
	return err
}

// Providers 返回支持的供应商枚举，供前端表单动态渲染。
func (s *ModelConfigService) Providers() []map[string]any {
	return []map[string]any{
		{
			"value": "openai", "label": "OpenAI 兼容",
			"hint":        "DeepSeek、通义千问、Kimi、智谱 GLM、OpenRouter、Ollama 等",
			"needBaseUrl": true, "baseURLPlaceholder": "如 https://api.deepseek.com/v1",
		},
		{"value": "ark", "label": "豆包（火山方舟 Ark）", "needBaseUrl": false},
		{"value": "gemini", "label": "Google Gemini", "needBaseUrl": false},
	}
}
