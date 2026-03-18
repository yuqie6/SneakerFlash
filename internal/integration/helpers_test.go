//go:build integration

package integration

import (
	"SneakerFlash/internal/db"
	"SneakerFlash/internal/testutil"
	"testing"

	"gorm.io/gorm"
)

func setupIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()

	testutil.SetupTestConfig()
	db.DB = testutil.NewSQLiteDB(t)
	testutil.SetupTestRedis(t)
	return db.DB
}
