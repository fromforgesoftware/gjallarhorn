//go:build integration

package internaltest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/fromforgesoftware/go-kit/migrator"
	"github.com/fromforgesoftware/go-kit/persistence/gormdb"
	"github.com/fromforgesoftware/go-kit/persistence/gormdb/gormdbtest"
	"github.com/stretchr/testify/require"
)

// GetDB returns a per-process singleton Postgres with Gjallarhorn's migrations
// applied by the real kit migrator. DB_SCHEMA=gjallarhorn mirrors prod; the
// common-pre-migration bootstrap creates the gjallarhorn schema before
// golang-migrate's tracking table needs it.
func GetDB(t *testing.T) *gormdb.DBClient {
	t.Helper()

	tdb := gormdbtest.GetDB(t, "gjallarhorn")
	if tdb == nil {
		t.Skip("test database unavailable (docker/gnomock); skipping integration test")
	}

	t.Setenv("DB_HOST", tdb.Host)
	t.Setenv("DB_PORT", fmt.Sprintf("%d", tdb.Port))
	t.Setenv("DB_USER", tdb.User)
	t.Setenv("DB_PASSWORD", tdb.Password)
	t.Setenv("DB_NAME", tdb.DBName)
	t.Setenv("DB_SSL", "disable")
	t.Setenv("DB_SCHEMA", "gjallarhorn")

	require.NoError(t, migrator.Up(context.Background(), os.DirFS(migratorDir()), migrator.WithServiceName("gjallarhorn")))
	return tdb.DBClient
}

// TruncateTables wipes Gjallarhorn's tables between tests sharing the singleton
// container.
func TruncateTables(t *testing.T, db *gormdb.DBClient) {
	t.Helper()
	require.NoError(t, db.Exec(`TRUNCATE TABLE gjallarhorn.notification, gjallarhorn.delivery_attempt, gjallarhorn.notification_template, gjallarhorn.notification_preference RESTART IDENTITY CASCADE;`).Error)
}

func migratorDir() string {
	_, f, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(f), "..", "..", "cmd", "migrator")
}
