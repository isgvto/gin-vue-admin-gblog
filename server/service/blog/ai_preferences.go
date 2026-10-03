package blog

import "strings"

var aiTones = map[string]string{
	"": "", "natural": "表达自然，保留作者原有文风", "formal": "语气正式、表述严谨", "friendly": "语气亲切，表达易懂",
	"technical":  "技术表达准确、术语一致，解释清楚，保留代码与技术结论，不添加未经证实的技术细节",
	"restrained": "语气克制客观，减少夸张修饰与空泛评价，不把推测改成确定事实",
}
var aiEditStrengths = map[string]string{
	"": "", "standard": "按当前任务优化措辞和句子衔接，保留作者原意与关键信息",
	"light": "只修正语病、歧义与不顺畅表达，尽量保留原句、用词和作者个人表达",
	"deep":  "在各段内部充分调整句式和表达，减少重复，改善逻辑衔接；不改变原意，不增删事实，不改变段落数量与顺序",
}
var aiLengths = map[string]string{
	"": "", "original": "保持与原文相近的篇幅", "shorter": "压缩冗余表达，篇幅约为原文的七成，保留关键信息", "longer": "适度补充解释，篇幅约为原文的一点三倍，不编造事实",
}

func writingPreferences(req *AiChatRequest) string {
	switch req.Action {
	case aiActionPolish, aiActionRewrite, aiActionContinue, aiActionCustom, aiActionChapter:
	default:
		return ""
	}
	var preferences []string
	if tone := aiTones[req.Tone]; tone != "" {
		preferences = append(preferences, tone)
	}
	if req.Action == aiActionPolish || req.Action == aiActionRewrite || req.Action == aiActionCustom {
		if strength := aiEditStrengths[req.EditStrength]; strength != "" {
			preferences = append(preferences, "修改力度："+strength)
		}
	}
	length := aiLengths[req.Length]
	if req.Action == aiActionChapter && strings.TrimSpace(req.ChapterDraft) == "" {
		length = map[string]string{"shorter": "本章约300至500字", "longer": "本章约1000至1500字", "original": "本章约600至1000字"}[req.Length]
	}
	// 续写生成新内容，没有可按比例缩放的原文，始终使用任务的300字上限。
	if length != "" && req.Action != aiActionContinue {
		preferences = append(preferences, length)
	}
	if len(preferences) == 0 {
		return ""
	}
	if req.Action == aiActionChapter {
		return "\n本章写作偏好（服从作者明确指令）：" + strings.Join(preferences, "；")
	}
	return "\n写作偏好（服从作者明确指令，续写仍不超过300字）：" + strings.Join(preferences, "；")
}
