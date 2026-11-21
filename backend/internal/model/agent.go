package model

import "time"

const (
	AgentRunQueued    = "queued"
	AgentRunRunning   = "running"
	AgentRunSucceeded = "succeeded"
	AgentRunFailed    = "failed"
	AgentRunCancelled = "cancelled"
)

type AgentSession struct {
	ID        string    `gorm:"column:id;primaryKey"`
	ProjectID string    `gorm:"column:project_id;not null;index"`
	Title     string    `gorm:"column:title;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (AgentSession) TableName() string { return "agent_sessions" }

type AgentProposal struct {
	Explanation      string            `json:"explanation"`
	ExpectedRevision int64             `json:"expectedRevision"`
	Operations       []CanvasOperation `json:"operations"`
}

type AgentRun struct {
	ID                string        `gorm:"column:id;primaryKey"`
	ProjectID         string        `gorm:"column:project_id;not null;index"`
	SessionID         string        `gorm:"column:session_id;not null;index"`
	Status            string        `gorm:"column:status;not null;index"`
	Message           string        `gorm:"column:message;not null"`
	SelectedNodesJSON []byte        `gorm:"column:selected_nodes_json;type:jsonb;not null"`
	ExpectedRevision  int64         `gorm:"column:expected_revision;not null"`
	Explanation       string        `gorm:"column:explanation;not null"`
	OperationsJSON    []byte        `gorm:"column:operations_json;type:jsonb;not null"`
	ErrorMessage      string        `gorm:"column:error_message;not null"`
	CreatedAt         time.Time     `gorm:"column:created_at;not null"`
	UpdatedAt         time.Time     `gorm:"column:updated_at;not null"`
	Proposal          AgentProposal `gorm:"-"`
}

func (AgentRun) TableName() string { return "agent_runs" }
