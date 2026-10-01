package blog

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
)

func TestArticleContextDistributedAndBounded(t *testing.T) {
	source := "开头" + strings.Repeat("甲", 10000) + "中部" + strings.Repeat("乙", 10000) + "结尾😀"
	text, info := articleContext(source, 8000)
	for _, part := range []string{"开头", "中部", "结尾😀", "[中间内容省略]"} {
		if !strings.Contains(text, part) {
			t.Fatal("missing", part)
		}
	}
	if utf8.RuneCountInString(text) > 8000 || !info.Truncated || info.Notice == "" {
		t.Fatal(info)
	}
	for _, budget := range []int{0, 1, 10, 30, 100} {
		text, _ := articleContext(source, budget)
		if utf8.RuneCountInString(text) > budget {
			t.Fatal(budget)
		}
	}
	text, info = articleContext("完整正文😀", 8000)
	if text != "完整正文😀" || info.Truncated {
		t.Fatal(text, info)
	}
}

func TestSelectionNeverSilentlyTruncated(t *testing.T) {
	old := global.GVA_CONFIG.AI.ContextLimit
	global.GVA_CONFIG.AI.ContextLimit = 100
	defer func() { global.GVA_CONFIG.AI.ContextLimit = old }()
	for _, action := range []string{"polish", "rewrite", "custom"} {
		req := &AiChatRequest{Action: action, Selection: strings.Repeat("字", 101), Instruction: "处理"}
		if err := validateAiChatRequest(req); err == nil {
			t.Fatal(action)
		}
	}
	req := &AiChatRequest{Action: "custom", Selection: "选区优先", Content: strings.Repeat("正文", 100), Instruction: "翻译选区"}
	msg := (&AiService{}).buildUserMessage(req)
	if !strings.Contains(msg, "选区优先") || !strings.Contains(msg, "翻译选区") {
		t.Fatal(msg)
	}
}

func TestCursorUTF16AndZero(t *testing.T) {
	offset := 0
	req := &AiChatRequest{Action: "continue", Content: "甲😀乙后文", CursorOffset: &offset}
	if err := validateAiChatRequest(req); err == nil {
		t.Fatal("zero cursor must not use document end")
	}
	offset = 3
	if got := cursorPrefix(req); got != "甲😀" {
		t.Fatal(got)
	}
	if err := validateAiChatRequest(req); err != nil {
		t.Fatal(err)
	}
	offset = 100
	if err := validateAiChatRequest(req); err == nil {
		t.Fatal("invalid offset")
	}
	req.CursorOffset = nil
	req.CursorContext = "旧版上下文"
	if got := cursorPrefix(req); got != "旧版上下文" {
		t.Fatal(got)
	}
}

func TestHistoryCharacterBudget(t *testing.T) {
	history := make([]AiChatMessage, 10)
	for i := range history {
		history[i] = AiChatMessage{Role: "user", Content: strings.Repeat("字", 8000)}
	}
	msgs := (&AiService{}).trimHistory(history)
	total := 0
	for _, msg := range msgs {
		total += utf8.RuneCountInString(msg.Content)
	}
	if total > 6000 {
		t.Fatal(total)
	}
}

func TestNormalizeTagSuggestion(t *testing.T) {
	categories := map[string]uint{"技术": 1}
	tags := map[string]uint{"Vue": 2, "Go": 3}
	result, err := normalizeTagSuggestion("```json\n"+`{"category":"幻觉分类","tags":[" vue ","Vue","Go","候选新标签"],"newTags":["go","新标签","重复上限"]}`+"\n```", categories, tags)
	if err != nil {
		t.Fatal(err)
	}
	if result.Category != "" || len(result.Warnings) == 0 || strings.Join(result.Tags, ",") != "Vue,Go" || len(result.NewTags) != 2 || result.TagIDs[0] != 2 {
		t.Fatalf("%+v", result)
	}
	result, err = normalizeTagSuggestion(`{"category":" 技术 ","tags":[],"newTags":["<script>","Vue"]}`, categories, tags)
	if err != nil || result.CategoryID != 1 || len(result.NewTags) != 0 || result.TagIDs[0] != 2 {
		t.Fatal(result, err)
	}
	for _, bad := range []string{`null`, `{"tags":"wrong","newTags":[]}`, `{"tags":[],"newTags":[]} trailing`, `{"tags":[],"newTags":[],"unknown":true}`, `{"category":"技术"}`} {
		if _, err := normalizeTagSuggestion(bad, categories, tags); err == nil {
			t.Fatal(bad)
		}
	}
}
