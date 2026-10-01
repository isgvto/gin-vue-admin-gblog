package blog

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
)

// ContextLimit 是正文字符预算，不等同于供应商的 token 上限。
func aiContextLimit() int {
	limit := global.GVA_CONFIG.AI.ContextLimit
	if limit <= 0 {
		return 8000
	}
	return min(limit, 200000)
}

type AiContextInfo struct {
	OriginalChars int    `json:"originalChars"`
	UsedChars     int    `json:"usedChars"`
	Truncated     bool   `json:"truncated"`
	Notice        string `json:"notice"`
}

// 长文选取均匀分布的五个窗口，包括开头、正文中部和结尾；省略处显式分隔。
// 保持确定性和一次模型调用，不将抽样结果描述为阅读全文。
func articleContext(text string, limit int) (string, AiContextInfo) {
	runes := []rune(text)
	info := AiContextInfo{OriginalChars: len(runes), UsedChars: len(runes)}
	if len(runes) <= limit {
		return text, info
	}
	info.Truncated = true
	marker := "\n\n[中间内容省略]\n\n"
	count := min(5, max(1, limit/(utf8.RuneCountInString(marker)+1)))
	available := max(0, limit-(count-1)*utf8.RuneCountInString(marker))
	var parts []string
	used := 0
	for i := 0; i < count; i++ {
		size := available / count
		if i < available%count {
			size++
		}
		start := 0
		if count > 1 {
			start = i * (len(runes) - size) / (count - 1)
		}
		parts = append(parts, string(runes[start:start+size]))
		used += size
	}
	info.UsedChars = used
	info.Notice = fmt.Sprintf("正文共 %d 字，本次选取 %d 字作为参考，包含分布于全文的片段；未读取的部分可能影响结果。", len(runes), used)
	return strings.Join(parts, marker), info
}

// 浏览器选区位置使用 UTF-16，不能直接作为 Go rune 或字节偏移。
func cursorPrefix(req *AiChatRequest) string {
	if req.CursorOffset == nil {
		if req.CursorContext != "" {
			return req.CursorContext
		}
		return req.Content
	}
	units := 0
	for index, r := range req.Content {
		size := 1
		if r > 0xffff {
			size = 2
		}
		if units+size > *req.CursorOffset {
			return req.Content[:index]
		}
		units += size
	}
	return req.Content
}

func chatContext(req *AiChatRequest) (string, AiContextInfo) {
	limit := aiContextLimit()
	switch req.Action {
	case aiActionPolish, aiActionRewrite:
		length := utf8.RuneCountInString(req.Selection)
		return req.Selection, AiContextInfo{OriginalChars: length, UsedChars: length}
	case aiActionContinue:
		runes := []rune(cursorPrefix(req))
		info := AiContextInfo{OriginalChars: len(runes), UsedChars: len(runes)}
		if len(runes) > limit {
			info.UsedChars = limit
			info.Truncated = true
			info.Notice = fmt.Sprintf("本次续写参考光标前最近 %d 字。", limit)
			return string(runes[len(runes)-limit:]), info
		}
		return string(runes), info
	case aiActionCustom:
		if strings.TrimSpace(req.Selection) != "" {
			budget := limit - utf8.RuneCountInString(req.Selection)
			background, info := articleContext(req.Content, max(0, budget))
			if budget <= 0 {
				info.Notice = "本次使用完整选区，未附加其他正文背景。"
			}
			return "当前选区（优先处理此内容）：\n" + req.Selection + "\n\n文章背景（仅供参考）：\n" + background, info
		}
	}
	source := req.Content
	if source == "" {
		source = req.Instruction
	}
	return articleContext(source, limit)
}

func (s *AiService) ContextInfo(req *AiChatRequest) AiContextInfo {
	_, info := chatContext(req)
	return info
}
