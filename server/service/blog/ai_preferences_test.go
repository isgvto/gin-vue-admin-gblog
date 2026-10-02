package blog

import (
	"strings"
	"testing"
)

func TestWritingPreferenceValidation(t *testing.T) {
	for _, req := range []*AiChatRequest{{Tone: "injected"}, {Length: "invalid"}, {EditStrength: "invalid"}} {
		if ValidateAiRequestSize(req) == nil {
			t.Fatal("invalid preferences accepted")
		}
	}
	if err := ValidateAiRequestSize(&AiChatRequest{Tone: "formal", Length: "shorter"}); err != nil {
		t.Fatal(err)
	}
}

func TestWritingPreferenceStrengthAndScope(t *testing.T) {
	svc := &AiService{}
	for _, tone := range []string{"technical", "restrained"} {
		for _, strength := range []string{"light", "standard", "deep"} {
			req := &AiChatRequest{Action: "polish", Selection: "原文", Tone: tone, Length: "longer", EditStrength: strength}
			if err := ValidateAiRequestSize(req); err != nil {
				t.Fatal(err)
			}
			message := svc.buildUserMessage(req)
			if !strings.Contains(message, aiTones[tone]) || !strings.Contains(message, aiEditStrengths[strength]) {
				t.Fatal(message)
			}
		}
	}
	for _, action := range []string{"outline", "title", "summary", "suggest-tags"} {
		if got := writingPreferences(&AiChatRequest{Action: action, Tone: "technical", Length: "longer", EditStrength: "deep"}); got != "" {
			t.Fatalf("preferences leaked into %s: %s", action, got)
		}
	}
	got := writingPreferences(&AiChatRequest{Action: "continue", Tone: "technical", Length: "longer", EditStrength: "deep"})
	if strings.Contains(got, "修改力度") || strings.Contains(got, "一点三倍") || !strings.Contains(got, "300字") || !strings.Contains(got, "技术表达") {
		t.Fatal(got)
	}
	got = writingPreferences(&AiChatRequest{Action: "chapter", Tone: "restrained", Length: "longer", EditStrength: "deep"})
	if strings.Contains(got, "修改力度") || !strings.Contains(got, "1000至1500字") {
		t.Fatal(got)
	}
	if err := ValidateAiRequestSize(&AiChatRequest{}); err != nil {
		t.Fatalf("legacy requests must remain valid: %v", err)
	}
}

func TestWritingPreferencesAppliedToBodyActions(t *testing.T) {
	svc := &AiService{}
	req := &AiChatRequest{Action: "rewrite", Selection: "原选区", Tone: "formal", Length: "shorter"}
	message := svc.buildUserMessage(req)
	for _, part := range []string{"原选区", "正式", "七成"} {
		if !strings.Contains(message, part) {
			t.Fatal(message)
		}
	}
	req.Action = "summary"
	if writingPreferences(req) != "" {
		t.Fatal("body preferences leaked into structured task")
	}
	req.Action = "custom"
	req.Instruction = "保持原篇幅"
	if !strings.Contains(svc.buildUserMessage(req), "服从作者明确指令") {
		t.Fatal("instruction precedence missing")
	}
}

func TestTitlePromptSupportsTitleOnly(t *testing.T) {
	req := &AiChatRequest{Action: "title", Title: "仅有主题的文章"}
	if err := validateAiChatRequest(req); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains((&AiService{}).buildUserMessage(req), req.Title) {
		t.Fatal("title-only topic omitted")
	}
}
