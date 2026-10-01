package request

import (
	"github.com/isgvto/gin-vue-admin-gblog/server/model/common/request"
)

type AiModelConfigSearch struct {
	Name     string `json:"name" form:"name"`
	Provider string `json:"provider" form:"provider"`
	request.PageInfo
}

type AiModelConfigUpsert struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Provider    string  `json:"provider"`
	BaseURL     string  `json:"baseUrl"`
	APIKey      string  `json:"apiKey"`
	Model       string  `json:"model"`
	Temperature float32 `json:"temperature"`
	MaxTokens   int     `json:"maxTokens"`
	Status      bool    `json:"status"`
	Remark      string  `json:"remark"`
}
