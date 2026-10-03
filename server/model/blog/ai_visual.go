package blog

import "time"

// Temporary data is private and expires after one hour. Adopted records retain
// the durable file reference for idempotency; binary data is then removed.
type AiVisualTask struct {
	ID             string     `gorm:"primaryKey;size:64" json:"id"`
	UserID         uint       `gorm:"index" json:"-"`
	Kind           string     `gorm:"size:24" json:"kind"`
	Prompt         string     `gorm:"type:text" json:"prompt"`
	Mermaid        string     `gorm:"type:text" json:"mermaid,omitempty"`
	PNG            []byte     `json:"-"`
	CreatedAt      time.Time  `json:"createdAt"`
	ExpiresAt      time.Time  `gorm:"index" json:"expiresAt"`
	UploadingUntil *time.Time `json:"-"`
	URL            string     `gorm:"size:1024" json:"url,omitempty"`
	FileID         uint       `json:"fileId,omitempty"`
}

func (AiVisualTask) TableName() string { return "blog_ai_visual_tasks" }
