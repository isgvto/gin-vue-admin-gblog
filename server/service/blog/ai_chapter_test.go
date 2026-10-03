package blog

import (
	"strings"
	"testing"

	blogReq "github.com/isgvto/gin-vue-admin-gblog/server/model/blog/request"
)

func chapterRequest() *AiChatRequest {
	index := 1
	return &AiChatRequest{Action: "chapter", Title: "文章主题", Content: "已经采纳的第一章", ChapterIndex: &index,
		Outline:      []blogReq.AiOutlineSection{{Title: "背景", Brief: "介绍问题"}, {Title: "实践", Brief: "给出示例"}},
		ChapterDraft: "第二章的旧草稿", Instruction: "补充实际案例", Tone: "formal", Length: "shorter"}
}

func TestChapterPromptAndValidation(t *testing.T) {
	req := chapterRequest()
	if err := validateAiChatRequest(req); err != nil {
		t.Fatal(err)
	}
	message := (&AiService{}).buildUserMessage(req)
	for _, expected := range []string{"文章主题", "第2章：实践", "介绍问题", "给出示例", "已经采纳的第一章", "第二章的旧草稿", "补充实际案例", "只输出本章", "正式", "七成"} {
		if !strings.Contains(message, expected) {
			t.Fatalf("missing %q in chapter prompt", expected)
		}
	}
	req.ChapterDraft = ""
	if !strings.Contains((&AiService{}).buildUserMessage(req), "300至500字") {
		t.Fatal("new chapter has no explicit length preference")
	}
}

func TestChapterRejectsInvalidOrOversizedInputs(t *testing.T) {
	for name, mutate := range map[string]func(*AiChatRequest){
		"missing outline":   func(r *AiChatRequest) { r.Outline = nil },
		"missing index":     func(r *AiChatRequest) { r.ChapterIndex = nil },
		"negative index":    func(r *AiChatRequest) { *r.ChapterIndex = -1 },
		"out of range":      func(r *AiChatRequest) { *r.ChapterIndex = 2 },
		"empty title":       func(r *AiChatRequest) { r.Outline[0].Title = " " },
		"multiline title":   func(r *AiChatRequest) { r.Outline[0].Title = "标题\n另一个" },
		"long title":        func(r *AiChatRequest) { r.Outline[0].Title = strings.Repeat("字", 121) },
		"long brief":        func(r *AiChatRequest) { r.Outline[0].Brief = strings.Repeat("字", 601) },
		"long draft":        func(r *AiChatRequest) { r.ChapterDraft = strings.Repeat("😀", 8001) },
		"selection":         func(r *AiChatRequest) { r.Selection = "意外选区" },
		"too many chapters": func(r *AiChatRequest) { r.Outline = make([]blogReq.AiOutlineSection, 21) },
		"large outline": func(r *AiChatRequest) {
			r.Outline = make([]blogReq.AiOutlineSection, 20)
			for i := range r.Outline {
				r.Outline[i] = blogReq.AiOutlineSection{Title: "章节", Brief: strings.Repeat("字", 500)}
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			req := chapterRequest()
			mutate(req)
			if validateAiChatRequest(req) == nil {
				t.Fatal("invalid chapter accepted")
			}
		})
	}
}
