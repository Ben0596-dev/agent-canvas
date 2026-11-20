package model

import "time"

type OperationRequest struct {
	ProjectID      string    `gorm:"column:project_id;primaryKey"`
	RequestID      string    `gorm:"column:request_id;primaryKey"`
	ResultRevision int64     `gorm:"column:result_revision;not null"`
	CreatedAt      time.Time `gorm:"column:created_at;not null"`
}

func (OperationRequest) TableName() string { return "canvas_operation_requests" }
