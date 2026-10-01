package ai

import (
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
)

// AiModelConfig 大模型配置，供模型工厂动态创建 ChatModel
type AiModelConfig struct {
	global.GVA_MODEL
	Name        string  `json:"name" gorm:"column:name;size:128;not null;comment:显示名"`
	Provider    string  `json:"provider" gorm:"column:provider;size:32;not null;comment:供应商 openai|ark|gemini"`
	BaseURL     string  `json:"baseUrl" gorm:"column:base_url;size:255;comment:OpenAI兼容端点"`
	APIKey      string  `json:"-" gorm:"column:api_key;size:255;comment:密钥"`
	Model       string  `json:"model" gorm:"column:model;size:128;not null;comment:模型标识"`
	Temperature float32 `json:"temperature" gorm:"column:temperature;default:0.7;comment:温度"`
	MaxTokens   int     `json:"maxTokens" gorm:"column:max_tokens;default:4096;comment:最大输出token"`
	IsDefault   bool    `json:"isDefault" gorm:"column:is_default;default:false;comment:默认模型"`
	Status      bool    `json:"status" gorm:"column:status;default:true;comment:启用"`
	Remark      string  `json:"remark" gorm:"column:remark;size:255;comment:备注"`
}

func (AiModelConfig) TableName() string {
	return "ai_model_config"
}
