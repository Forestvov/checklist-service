//go:build integration

package tasks_postgres_repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	postgres_testutil "github.com/Forestvov/checklist-service/internal/testutil/postgres"
)

var testDatabase *postgres_testutil.Database

const (
	testSetupTimeout     = 2 * time.Minute
	testOperationTimeout = 10 * time.Second
	testCleanupTimeout   = 5 * time.Second
	initialTaskVersion   = int64(1)
)

func TestMain(m *testing.M) {
	os.Exit(runIntegrationTests(m))
}

func runIntegrationTests(m *testing.M) (exitCode int) {
	setupCtx, cancel := context.WithTimeout(context.Background(), testSetupTimeout)
	database, err := postgres_testutil.Start(setupCtx)
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, "set up integration test PostgreSQL:", err)
		return 1
	}

	testDatabase = database
	defer func() {
		if err := database.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "tear down integration test PostgreSQL:", err)
			exitCode = 1
		}
	}()

	return m.Run()
}

func newTestRepository(t *testing.T) *TasksRepository {
	t.Helper()

	truncateTasks(t)
	t.Cleanup(func() {
		truncateTasks(t)
	})

	return NewTasksRepository(testDatabase.Pool)
}

func truncateTasks(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), testCleanupTimeout)
	defer cancel()

	const query = `TRUNCATE TABLE checklist.tasks RESTART IDENTITY CASCADE;`
	if _, err := testDatabase.Pool.Exec(ctx, query); err != nil {
		t.Fatalf("truncate tasks table: %v", err)
	}
}

func newTestContext(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(
		t.Context(),
		testOperationTimeout,
	)
	t.Cleanup(cancel)

	return ctx
}
