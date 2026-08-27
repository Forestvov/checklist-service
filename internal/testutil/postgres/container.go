//go:build integration

package postgres_testutil

import (
	"context"
	"errors"
	"fmt"
	"time"

	core_pgx_pool "github.com/Forestvov/checklist-service/internal/core/repository/postgres/pool/pgx"
	"github.com/testcontainers/testcontainers-go"
	postgres_container "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	postgresImage    = "postgres:18-bookworm"
	postgresDatabase = "checklist"
	postgresUser     = "checklist"
	postgresPassword = "checklist"
)

type Database struct {
	Pool      *core_pgx_pool.Pool
	container *postgres_container.PostgresContainer
}

func Start(ctx context.Context) (*Database, error) {
	container, err := postgres_container.Run(
		ctx,
		postgresImage,
		postgres_container.WithDatabase(postgresDatabase),
		postgres_container.WithUsername(postgresUser),
		postgres_container.WithPassword(postgresPassword),
		postgres_container.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, fmt.Errorf("start PostgreSQL container: %w", err)
	}

	database := &Database{container: container}

	host, err := container.Host(ctx)
	if err != nil {
		return closeAfterStartError(database, fmt.Errorf("get PostgreSQL host: %w", err))
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return closeAfterStartError(database, fmt.Errorf("get PostgreSQL port: %w", err))
	}

	pool, err := core_pgx_pool.NewPool(
		ctx,
		core_pgx_pool.Config{
			Host:     host,
			Port:     port.Port(),
			User:     postgresUser,
			Password: postgresPassword,
			Database: postgresDatabase,
			Timeout:  5 * time.Second,
		},
	)
	if err != nil {
		return closeAfterStartError(database, fmt.Errorf("connect to PostgreSQL: %w", err))
	}

	database.Pool = pool

	if err := applyUpMigrations(ctx, pool); err != nil {
		return closeAfterStartError(database, fmt.Errorf("apply migrations: %w", err))
	}

	return database, nil
}

func (d *Database) Close() error {
	if d == nil {
		return nil
	}

	if d.Pool != nil {
		d.Pool.Close()
	}

	if err := testcontainers.TerminateContainer(d.container); err != nil {
		return fmt.Errorf("terminate PostgreSQL container: %w", err)
	}

	return nil
}

func closeAfterStartError(database *Database, startErr error) (*Database, error) {
	if closeErr := database.Close(); closeErr != nil {
		return nil, errors.Join(startErr, closeErr)
	}

	return nil, startErr
}
