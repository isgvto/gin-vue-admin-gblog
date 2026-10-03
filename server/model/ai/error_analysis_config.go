package ai

// ErrorAnalysisConfig binds the error-analysis feature to an existing model.
// ModelID=0 explicitly selects the enabled database default, never YAML fallback.
type ErrorAnalysisConfig struct {
	ID                  uint `json:"-" gorm:"primaryKey"`
	PermissionsMigrated bool `json:"-"`
	Enabled             bool `json:"enabled"`
	ModelID             uint `json:"modelId"`
	TimeoutSeconds      int  `json:"timeoutSeconds"`
}

func (ErrorAnalysisConfig) TableName() string { return "ai_error_analysis_config" }

type ErrorAnalysisRequest struct {
	ID uint `json:"id" binding:"required"`
}
