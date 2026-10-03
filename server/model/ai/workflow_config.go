package ai

import "github.com/isgvto/gin-vue-admin-gblog/server/model/common"

// WorkflowConfig independently binds requirement analysis and prompt planning.
type WorkflowConfig ErrorAnalysisConfig

func (WorkflowConfig) TableName() string { return "ai_workflow_config" }

type WorkflowTurn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type WorkflowChatRequest struct {
	Mode           string         `json:"mode"`
	Query          string         `json:"query"`
	Inputs         common.JSONMap `json:"inputs"`
	History        []WorkflowTurn `json:"history"`
	SessionID      uint           `json:"sessionId"`
	ConversationID string         `json:"conversation_id"`
	ResponseMode   string         `json:"response_mode"`
}
