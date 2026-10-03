package blog

import (
	"fmt"
	"unicode/utf16"
	"unicode/utf8"
)

func ValidateAiChatRequest(req *AiChatRequest) error { return validateAiChatRequest(req) }

// 字符限制在构建提示词之前执行，HTTP 层另有 2 MiB 请求体限制。
func ValidateAiRequestSize(req *AiChatRequest) error {
	if _, ok := aiTones[req.Tone]; !ok {
		return fmt.Errorf("不支持的写作语气")
	}
	if _, ok := aiLengths[req.Length]; !ok {
		return fmt.Errorf("不支持的篇幅偏好")
	}
	if _, ok := aiEditStrengths[req.EditStrength]; !ok {
		return fmt.Errorf("不支持的修改力度")
	}
	if req.CursorOffset != nil && (*req.CursorOffset < 0 || *req.CursorOffset > len(utf16.Encode([]rune(req.Content)))) {
		return fmt.Errorf("光标位置无效")
	}
	if (req.Action == aiActionPolish || req.Action == aiActionRewrite || req.Action == aiActionCustom) && utf8.RuneCountInString(req.Selection) > aiContextLimit() {
		return fmt.Errorf("选区超过本次 %d 字处理上限，请分段选择；原文未作修改", aiContextLimit())
	}
	for _, field := range []struct {
		name, value string
		limit       int
	}{
		{"正文", req.Content, 200000}, {"选区", req.Selection, 200000},
		{"光标上下文", req.CursorContext, 200000}, {"标题", req.Title, 500},
		{"指令", req.Instruction, 8000},
		{"章节草稿", req.ChapterDraft, 8000},
	} {
		if utf8.RuneCountInString(field.value) > field.limit {
			return fmt.Errorf("%s超过 %d 字限制", field.name, field.limit)
		}
	}
	if err := validateOutlineSize(req.Outline); err != nil {
		return err
	}
	if len(req.History) > 20 {
		return fmt.Errorf("对话历史过长")
	}
	for _, msg := range req.History {
		if utf8.RuneCountInString(msg.Content) > 8000 {
			return fmt.Errorf("单条对话历史超过 8000 字限制")
		}
	}
	return nil
}
