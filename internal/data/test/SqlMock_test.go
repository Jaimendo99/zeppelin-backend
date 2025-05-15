package test_test

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres" // Use the GORM driver matching your actual DB
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"strings"
	"testing"
)

// setupMockDb initializes GORM with sqlmock
func setupMockDb(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn: db,
		//PreferSimpleProtocol: true, // Might be needed depending on query types/driver
	})

	gormDb, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	return gormDb, mock
}

func quoteSql(sql string) string {
	replacer := strings.NewReplacer(
		`.`, `\.`,
		`*`, `\*`,
		`+`, `\+`,
		`?`, `\?`,
		`|`, `\|`,
		`(`, `\(`,
		`)`, `\)`,
		`[`, `\[`,
		`{`, `\{`,
		`\`, `\\`,
		`^`, `\^`,
		`$`, `\$`,
	)
	return replacer.Replace(sql)
}
