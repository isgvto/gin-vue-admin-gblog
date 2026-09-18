package blog

import "strings"

var aiTones = map[string]string{
	"": "", "natural": "表达自然，保留作者原有文风", "formal": "语气正式、表述严谨", "friendly": "语气亲切，表达易懂",
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
	length := aiLengths[req.Length]
	if req.Action == aiActionChapter && strings.TrimSpace(req.ChapterDraft) == "" {
		length = map[string]string{"shorter": "本章约300至500字", "longer": "本章约1000至1500字", "original": "本章约600至1000字"}[req.Length]
	}
	if length != "" {
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
