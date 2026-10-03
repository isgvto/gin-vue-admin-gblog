package ai

// ImageConfig stores independent image credentials; ModelID is migration-only.
type ImageConfig struct {
	ID                  uint   `gorm:"primaryKey" json:"-"`
	Enabled             bool   `json:"enabled"`
	ModelID             uint   `json:"-"` // Legacy binding, used only by migration.
	Provider            string `json:"provider" gorm:"size:32"`
	BaseURL             string `json:"baseUrl" gorm:"size:255"`
	Model               string `json:"model" gorm:"size:128"`
	APIKey              string `json:"-" gorm:"size:255"`
	IndependentMigrated bool   `json:"-"`
	TestAccessMigrated  bool   `json:"-"`
	TimeoutSeconds      int    `json:"timeoutSeconds"`
	PermissionsMigrated bool   `json:"-"`
}

func (ImageConfig) TableName() string { return "ai_image_config" }
