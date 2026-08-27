//go:build integration

package tasks_postgres_repository

import (
	"errors"
	"testing"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

func TestTasksRepositoryDeleteTaskSuccess(t *testing.T) {
	repository := newTestRepository(t)
	ctx := newTestContext(t)

	description := "Milk, bread and vegetables"
	input := core_domain.NewTaskUninitialized(
		"Buy groceries",
		&description,
	)

	created, err := repository.CreateTask(ctx, input)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	err = repository.DeleteTask(ctx, created.ID)
	if err != nil {
		t.Fatalf("delete task: %v", err)
	}

	actual, err := repository.GetTask(ctx, created.ID)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf(
			"get deleted task: got error %v, want ErrNotFound",
			err,
		)
	}

	if actual != (core_domain.Task{}) {
		t.Errorf(
			"unexpected deleted task: got %+v, want zero task",
			actual,
		)
	}
}

func TestTasksRepositoryDeleteTaskNotFound(t *testing.T) {
	const missingTaskID int64 = 9999

	repository := newTestRepository(t)
	ctx := newTestContext(t)

	err := repository.DeleteTask(ctx, missingTaskID)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf(
			"unexpected error: got %v, want ErrNotFound",
			err,
		)
	}
}
