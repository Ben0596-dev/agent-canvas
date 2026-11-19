package db

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Open creates the persistence adapter. Schema changes are applied by the
// migration runner, never by GORM AutoMigrate.
func Open(ctx context.Context, databaseURL string) (*gorm.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is empty")
	}
	return gorm.Open(postgres.Open(databaseURL), &gorm.Config{TranslateError: true})
}
