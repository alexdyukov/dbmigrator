package main_test

import (
	"database/sql"
	"embed"
	"testing"

	"github.com/alexdyukov/dbmigrator"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

//go:embed *.sql
var embeded embed.FS

func TestPostgresql(t *testing.T) {
	t.Parallel()

	container, err := postgres.Run(t.Context(),
		"postgres",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.BasicWaitStrategies(),
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

	db, err := sql.Open("pgx", dsn)
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
