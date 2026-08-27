//go:build integration

package postgres_testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	core_pgx_pool "github.com/Forestvov/checklist-service/internal/core/repository/postgres/pool/pgx"
)

func applyUpMigrations(ctx context.Context, pool *core_pgx_pool.Pool) error {
	migrationPaths, err := findUpMigrations()
	if err != nil {
		return err
	}

	for _, migrationPath := range migrationPaths {
		migration, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", filepath.Base(migrationPath), err)
		}

		if _, err := pool.Exec(ctx, string(migration)); err != nil {
			return fmt.Errorf("execute migration %s: %w", filepath.Base(migrationPath), err)
		}
	}

	return nil
}

func findUpMigrations() ([]string, error) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("resolve test utility source path")
	}

	projectRoot := filepath.Clean(
		filepath.Join(filepath.Dir(sourceFile), "../../.."),
	)
	migrationPaths, err := filepath.Glob(
		filepath.Join(projectRoot, "migrations", "*.up.sql"),
	)
	if err != nil {
		return nil, fmt.Errorf("find up migrations: %w", err)
	}
	if len(migrationPaths) == 0 {
		return nil, fmt.Errorf("no up migrations found")
	}

	sort.Strings(migrationPaths)

	return migrationPaths, nil
}
