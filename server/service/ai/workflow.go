package ai

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	aiModel "github.com/isgvto/gin-vue-admin-gblog/server/model/ai"
	"github.com/isgvto/gin-vue-admin-gblog/server/model/system"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WorkflowService struct{}

func (s WorkflowService) Config() (aiModel.WorkflowConfig, error) {
	cfg := aiModel.WorkflowConfig{ID: 1, TimeoutSeconds: 180}
	err := global.GVA_DB.First(&cfg, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return aiModel.WorkflowConfig{ID: 1, TimeoutSeconds: 180}, nil
	}
	return cfg, err
}

func (s WorkflowService) Resolve(cfg aiModel.WorkflowConfig) (*aiModel.AiModelConfig, error) {
	if cfg.TimeoutSeconds < 10 || cfg.TimeoutSeconds > 600 {
		return nil, errors.New("工作流超时须在 10～600 秒之间")
	}
	return resolveFeatureModel(cfg.ModelID, "AI 需求工作流")
}

func (s WorkflowService) Save(cfg aiModel.WorkflowConfig) error {
	if cfg.TimeoutSeconds < 10 || cfg.TimeoutSeconds > 600 {
		return errors.New("工作流超时须在 10～600 秒之间")
	}
	if cfg.Enabled {
		if _, err := s.Resolve(cfg); err != nil {
			return err
		}
	}
	cfg.ID = 1
	return global.GVA_DB.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.AssignmentColumns([]string{"enabled", "model_id", "timeout_seconds"})}).Create(&cfg).Error
}

type WorkflowRun struct {
	Config         aiModel.WorkflowConfig
	Model          *aiModel.AiModelConfig
	Messages       []*schema.Message
	Mode           string
	ConversationID string
	MessageID      string
}

func newWorkflowID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return "local-" + hex.EncodeToString(value), nil
}

func (s WorkflowService) Prepare(ctx context.Context, userID uint, req aiModel.WorkflowChatRequest) (*WorkflowRun, error) {
	if userID == 0 {
		return nil, errors.New("用户未登录")
	}
	messages, err := workflowMessages(req)
	if err != nil {
		return nil, err
	}
	if req.SessionID != 0 {
		var session system.SysAIWorkflowSession
		if err := global.GVA_DB.WithContext(ctx).Select("id", "tab").Where("id = ? AND user_id = ?", req.SessionID, userID).First(&session).Error; err != nil {
			return nil, errors.New("会话不存在或无权访问")
		}
		tab := "analysis"
		if req.Mode == "workflowPromptChat" {
			tab = "workflow"
		}
		if session.Tab != tab {
			return nil, errors.New("会话类型与请求模式不一致")
		}
	}
	cfg, err := s.Config()
	if err != nil {
		return nil, err
	}
	if !cfg.Enabled {
		return nil, errors.New("AI 需求工作流未启用，请在 AI 模型配置中启用并分配模型")
	}
	model, err := s.Resolve(cfg)
	if err != nil {
		return nil, err
	}
	conversationID := req.ConversationID
	if !strings.HasPrefix(conversationID, "local-") || len(conversationID) > 80 {
		conversationID, err = newWorkflowID()
		if err != nil {
			return nil, err
		}
	}
	messageID, err := newWorkflowID()
	if err != nil {
		return nil, err
	}
	for _, msg := range messages {
		msg.Content = strings.ReplaceAll(msg.Content, model.APIKey, "[已脱敏]")
	}
	return &WorkflowRun{Config: cfg, Model: model, Messages: messages, Mode: req.Mode, ConversationID: conversationID, MessageID: messageID}, nil
}

func (r *WorkflowRun) Stream(ctx context.Context) (*schema.StreamReader[*schema.Message], error) {
	model, err := Factory().build(ctx, r.Model)
	if err != nil {
		return nil, r.SafeError(err)
	}
	stream, err := model.Stream(ctx, r.Messages)
	if err != nil {
		return nil, r.SafeError(err)
	}
	return stream, nil
}

func (r *WorkflowRun) Generate(ctx context.Context) (string, error) {
	model, err := Factory().build(ctx, r.Model)
	if err != nil {
		return "", r.SafeError(err)
	}
	reply, err := model.Generate(ctx, r.Messages)
	if err != nil {
		return "", r.SafeError(err)
	}
	if reply == nil {
		return "", errors.New("模型未返回有效内容")
	}
	text := strings.ReplaceAll(reply.Content, r.Model.APIKey, "[已脱敏]")
	result, err := ValidateWorkflowResult(r.Mode, text)
	if err != nil {
		return "", err
	}
	canonical, err := json.MarshalIndent(result, "", "  ")
	return string(canonical), err
}

func (r *WorkflowRun) SafeError(err error) error { return analysisError(err, r.Model.APIKey) }

func (s WorkflowService) Test(ctx context.Context, cfg aiModel.WorkflowConfig) (string, string, error) {
	model, err := s.Resolve(cfg)
	if err != nil {
		return "", "", err
	}
	messages, err := workflowMessages(aiModel.WorkflowChatRequest{Mode: "analysisChat", Query: "设计一个分类管理功能，包含名称、排序和启用状态，并支持基础增删改查。"})
	if err != nil {
		return "", "", err
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.TimeoutSeconds)*time.Second)
	defer cancel()
	run := &WorkflowRun{Model: model, Messages: messages, Mode: "analysisChat"}
	text, err := run.Generate(ctx)
	return text, model.Name + " / " + model.Model, err
}

const workflowAnalysisSchema = `{"summary":"分析摘要","recommendedPackageType":"普通模块或插件建议","missingInfo":["待确认问题"],"suggestions":["建议"],"modules":[{"name":"category","label":"分类","description":"模块说明","fields":[{"name":"name","label":"名称","type":"string","required":true,"description":"字段说明","dictionary":"","relation":""}]}],"clientPages":[{"name":"categoryList","label":"分类列表","description":"页面说明","pageType":"list","targetModules":["category"],"fields":[{"name":"name","label":"名称","sourceModule":"category","sourceField":"name","displayType":"input","required":true,"description":""}],"interactions":["交互"],"relations":["映射"]}]}`
const workflowStepsSchema = `{"summary":"执行路线摘要","steps":[{"title":"步骤标题","goal":"步骤目标","prompt":"可复制的详细提示词","expectedOutput":"预期产物与验收标准","suggestedTool":"建议使用的现有工具","autoExecutable":false}]}`

func workflowMessages(req aiModel.WorkflowChatRequest) ([]*schema.Message, error) {
	if req.Mode != "analysisChat" && req.Mode != "workflowPromptChat" {
		return nil, errors.New("不支持的工作流模式")
	}
	if strings.TrimSpace(req.Query) == "" || len([]rune(req.Query)) > 6000 {
		return nil, errors.New("请输入 1～6000 字的需求或追问")
	}
	if len(req.History) > 12 {
		return nil, errors.New("会话上下文最多 12 条消息")
	}
	inputs, err := json.Marshal(req.Inputs)
	if err != nil || len(inputs) > 100000 {
		return nil, errors.New("工作流输入过长或格式无效")
	}
	schemaText := workflowAnalysisSchema
	goal := "分析需求，梳理模块、字段、字典、关联和待确认项；输入要求客户端页面时补充页面和后端字段映射，不需要时 clientPages 为空数组。"
	if req.Mode == "workflowPromptChat" {
		schemaText = workflowStepsSchema
		goal = "根据需求、已有分析、flowType 和额外约束生成可复制的分步骤 Prompt，涵盖配置、预览、实现、编译、权限和验证。"
	}
	prompt := "你是本项目 Go/Gin/GORM + Vue/Element Plus 的开发规划助手。用中文。" + goal + "已有工具包括自动代码生成、模板、表单设计、插件、MCP。当前只提供分析和建议，不执行代码生成或安装操作，不声称步骤已执行，autoExecutable 始终为 false。不输出凭据。返回一个完整 JSON 对象，禁止思考标签、Markdown 围栏或对象外文字。字段类型严格遵守以下结构（示例内容须替换，数组可为空）：" + schemaText
	messages := []*schema.Message{schema.SystemMessage(prompt)}
	total := 0
	for _, turn := range req.History {
		if turn.Role != "user" && turn.Role != "assistant" {
			return nil, errors.New("上下文仅支持 user 和 assistant 消息")
		}
		size := len([]rune(turn.Content))
		total += size
		if size > 16000 || total > 80000 {
			return nil, errors.New("会话上下文过长，请新建会话或回滚到较近节点")
		}
		text := SanitizeAnalysisText(turn.Content)
		if turn.Role == "user" {
			messages = append(messages, schema.UserMessage(text))
		} else {
			messages = append(messages, schema.AssistantMessage(text, nil))
		}
	}
	messages = append(messages, schema.UserMessage(fmt.Sprintf("本轮问题：\n%s\n\n本轮配置：\n%s", SanitizeAnalysisText(req.Query), SanitizeAnalysisText(string(inputs)))))
	return messages, nil
}

// ValidateWorkflowResult rejects malformed structures before they reach UI normalizers.
func ValidateWorkflowResult(mode, text string) (map[string]any, error) {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```json") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimSuffix(strings.TrimSpace(text), "```")
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(text), &result); err != nil || result == nil {
		return nil, errors.New("模型未返回有效 JSON，请重试或更换模型")
	}
	if summary, ok := result["summary"].(string); !ok || strings.TrimSpace(summary) == "" {
		return nil, errors.New("模型结果缺少有效摘要")
	}
	array := func(parent map[string]any, key string) ([]any, error) {
		if parent[key] == nil {
			return nil, nil
		}
		list, ok := parent[key].([]any)
		if !ok {
			return nil, fmt.Errorf("模型结果 %s 必须是数组", key)
		}
		return list, nil
	}
	if mode == "workflowPromptChat" {
		steps, err := array(result, "steps")
		if err != nil {
			return nil, err
		}
		if len(steps) == 0 {
			return nil, errors.New("模型结果未提供工作流步骤")
		}
		for _, value := range steps {
			step, ok := value.(map[string]any)
			if !ok {
				return nil, errors.New("模型步骤格式无效")
			}
			if p, ok := step["prompt"].(string); !ok || strings.TrimSpace(p) == "" {
				return nil, errors.New("模型步骤缺少提示词")
			}
			step["autoExecutable"] = false
		}
	} else {
		if _, exists := result["modules"]; !exists {
			return nil, errors.New("模型分析结果缺少模块列表")
		}
		for _, key := range []string{"modules", "clientPages"} {
			items, err := array(result, key)
			if err != nil {
				return nil, err
			}
			for _, value := range items {
				item, ok := value.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("模型结果 %s 项格式无效", key)
				}
				fields, err := array(item, "fields")
				if err != nil {
					return nil, err
				}
				for _, field := range fields {
					if _, ok := field.(map[string]any); !ok {
						return nil, errors.New("模型字段格式无效")
					}
				}
			}
		}
	}
	return result, nil
}
