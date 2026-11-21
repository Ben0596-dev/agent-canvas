package model

import "time"

const (
	GenerationKindImage = "image"
	GenerationKindVideo = "video"

	GenerationTaskQueued    = "queued"
	GenerationTaskRunning   = "running"
	GenerationTaskSucceeded = "succeeded"
	GenerationTaskFailed    = "failed"
)

type GenerationTask struct {
	ID             string    `gorm:"column:id;primaryKey"`
	ProjectID      string    `gorm:"column:project_id;not null;index"`
	Kind           string    `gorm:"column:kind;not null"`
	Provider       string    `gorm:"column:provider;not null"`
	ModelID        string    `gorm:"column:model_id;not null"`
	UpstreamModel  string    `gorm:"column:upstream_model;not null"`
	Status         string    `gorm:"column:status;not null;index"`
	Prompt         string    `gorm:"column:prompt;not null"`
	ParametersJSON []byte    `gorm:"column:parameters_json;type:jsonb;not null"`
	ProviderTaskID string    `gorm:"column:provider_task_id;not null"`
	ResultURL      string    `gorm:"column:result_url;not null"`
	CoverURL       string    `gorm:"column:cover_url;not null"`
	ErrorMessage   string    `gorm:"column:error_message;not null"`
	CreatedAt      time.Time `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null"`
}

func (GenerationTask) TableName() string { return "generation_tasks" }

type ModelProfile struct {
	ID            string
	Provider      string
	UpstreamModel string
	Capability    string
	Defaults      map[string]any
}
