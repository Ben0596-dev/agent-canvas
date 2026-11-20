package model

import "time"

type Project struct {
	ID           string    `gorm:"column:id;primaryKey"`
	UserID       string    `gorm:"column:user_id;not null;index:idx_canvas_projects_user_updated,priority:1"`
	Title        string    `gorm:"column:title;not null"`
	Revision     int64     `gorm:"column:revision;not null"`
	SnapshotJSON []byte    `gorm:"column:snapshot_json;type:jsonb;not null"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;index:idx_canvas_projects_user_updated,priority:2,sort:desc"`
}

func (Project) TableName() string { return "canvas_projects" }
