// Package dbmigrator provides method for database migration any stdlib's sql.DB compatibility database
package dbmigrator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"sort"
)

type (
	invalidDirectoryStructureError struct {
		entryName string
	}
	// DBPool is the interface that describes minimal requirements of database connection.
	DBPool interface {
		BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	}
)

func (err invalidDirectoryStructureError) Error() string {
	return "directories in migration folder not allowed: " + err.entryName
}

func parseMigrations(fsys fs.FS) (map[string]string, []string, error) {
	fileInfos, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list directory entries with error: %w", err)
	}

	migrations := make(map[string]string, len(fileInfos))
	upgradePlan := make([]string, 0, len(fileInfos))

	for _, entry := range fileInfos {
		entryName := entry.Name()

		if entry.IsDir() {
			return nil, nil, invalidDirectoryStructureError{entryName: entryName}
		}

		body, err := fs.ReadFile(fsys, entryName)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read file %s with error: %w", entryName, err)
		}

		migrations[entryName] = string(body)

		upgradePlan = append(upgradePlan, entryName)
	}

	sort.Strings(upgradePlan)

	return migrations, upgradePlan, nil
}

func initializeVersionScheme(ctx context.Context, pool DBPool, versionSchemeName string) (err error) {
	var transaction *sql.Tx

	if transaction, err = pool.BeginTx(ctx, nil); err != nil {
		err = fmt.Errorf("failed to begin transaction with error: %w", err)

		return
	}

	defer func() {
		if err == nil {
			err = transaction.Commit()

			return
		}

		if rollbackErr := transaction.Rollback(); rollbackErr != nil {
			err = fmt.Errorf("failed to rollback for base error: %w", err)
		}
	}()

	cmd := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (version VARCHAR(255) UNIQUE NOT NULL);", versionSchemeName)
	if _, err = transaction.ExecContext(ctx, cmd); err != nil {
		err = fmt.Errorf("failed to initialize version scheme with error: %w", err)
	}

	return
}

func migrateOne(ctx context.Context, pool DBPool, versionSchemeName, version, migrationCommand string) (err error) {
	var (
		transaction *sql.Tx
		cmd         string
	)

	if transaction, err = pool.BeginTx(ctx, nil); err != nil {
		err = fmt.Errorf("failed to begin transaction with error: %w", err)

		return err
	}

	defer func() {
		if err == nil {
			err = transaction.Commit()

			return
		}

		if rollbackErr := transaction.Rollback(); rollbackErr != nil {
			err = fmt.Errorf("failed to rollback for base error: %w", err)
		}
	}()

	cmd = fmt.Sprintf("SELECT version FROM %s WHERE version='%s';", versionSchemeName, version)
	if err = transaction.QueryRowContext(ctx, cmd).Scan(&cmd); err == nil {
		// already migrated
		return
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		err = fmt.Errorf("failed to get version from version scheme with error: %w", err)

		return
	}

	cmd = migrationCommand
	if _, err = transaction.ExecContext(ctx, cmd); err != nil {
		err = fmt.Errorf("failed to execute migration %s with error: %w", version, err)

		return
	}

	cmd = fmt.Sprintf("INSERT INTO %s VALUES ('%s');", versionSchemeName, version)
	if _, err = transaction.ExecContext(ctx, cmd); err != nil {
		err = fmt.Errorf("failed to insert migration into version scheme with error: %w", err)
	}

	return
}

// Migrate parses migration from fs.FS and run them one by one in a lexical sort manner.
func Migrate(ctx context.Context, fsys fs.FS, pool DBPool, versionSchemeName string) error {
	migrations, upgradePlan, err := parseMigrations(fsys)
	if err != nil {
		return fmt.Errorf("dbmigrator: %w", err)
	}

	if err = initializeVersionScheme(ctx, pool, versionSchemeName); err != nil {
		return fmt.Errorf("dbmigrator: %w", err)
	}

	for i := 0; i < len(upgradePlan); i++ {
		version := upgradePlan[i]

		if err := migrateOne(ctx, pool, versionSchemeName, version, migrations[version]); err != nil {
			return fmt.Errorf("dbmigrator: %w", err)
		}
	}

	return nil
}
