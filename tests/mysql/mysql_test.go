package main_test

import (
	"database/sql"
	"embed"
	"testing"

	"github.com/alexdyukov/dbmigrator"
	_ "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
)

//go:embed *.sql
var embeded embed.FS

func TestMysql(t *testing.T) {
	t.Parallel()

	container, err := mysql.Run(t.Context(),
		"mysql",
		mysql.WithDatabase("testdb"),
		mysql.WithUsername("testuser"),
		mysql.WithPassword("testpass"),
	)
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}

	defer func() {
		err = testcontainers.TerminateContainer(container)
		if err != nil {
			t.Fatalf("failed to terminate container: %v", err)
		}
	}()

	dsn, err := container.ConnectionString(t.Context())
	if err != nil {
		t.Fatalf("failed to create connection string: %v", err)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("failed to initialize sql.DB: %v", err)
	}

	if err = dbmigrator.Migrate(t.Context(), embeded, db, "migrations"); err != nil {
		t.Fatalf("failed to init migrate: %v", err)
	}

	if err = dbmigrator.Migrate(t.Context(), embeded, db, "migrations"); err != nil {
		t.Fatalf("failed to migrate already migrated: %v", err)
	}
}
