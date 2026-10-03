package request

type AiChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AiOutlineSection struct {
	Title string `json:"title"`
	Brief string `json:"brief"`
}

type AiChatRequest struct {
	Action        string             `json:"action"`
	Tone          string             `json:"tone,omitempty"`
	Length        string             `json:"length,omitempty"`
	EditStrength  string             `json:"editStrength,omitempty"`
	Content       string             `json:"content"`
	Selection     string             `json:"selection"`
	CursorContext string             `json:"cursorContext"`
	CursorOffset  *int               `json:"cursorOffset,omitempty"`
	Title         string             `json:"title"`
	Instruction   string             `json:"instruction"`
	Catalog       string             `json:"catalog,omitempty"`
	History       []AiChatMessage    `json:"history"`
	Outline       []AiOutlineSection `json:"outline,omitempty"`
	ChapterIndex  *int               `json:"chapterIndex,omitempty"`
	ChapterDraft  string             `json:"chapterDraft,omitempty"`
}
