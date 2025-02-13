package db

import (
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB      *gorm.DB
	DbError error
	Once    sync.Once
)

func InitDb(connectionString string) error {
	Once.Do(func() {
		DB, DbError = gorm.Open(postgres.New(postgres.Config{
			DSN:                  connectionString,
			PreferSimpleProtocol: true,
		}), &gorm.Config{
			TranslateError: true,
			Logger:         logger.Default.LogMode(logger.Info),
		})
	})
	return DbError
}
