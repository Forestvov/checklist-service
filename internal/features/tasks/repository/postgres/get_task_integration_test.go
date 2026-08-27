//go:build integration

package tasks_postgres_repository

import (
	"errors"
	"testing"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

func TestTasksRepositoryGetTaskSuccess(t *testing.T) {
	repository := newTestRepository(t)
	ctx := newTestContext(t)

	description := "Milk, bread and vegetables"
	input := core_domain.NewTaskUninitialized(
		"Buy groceries",
		&description,
	)

	expected, err := repository.CreateTask(ctx, input)
	if err != nil {
		t.Fatalf("prepare task: %v", err)
	}

	actual, err := repository.GetTask(ctx, expected.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}

	if actual.ID != expected.ID {
		t.Errorf("unexpected ID: got %d, want %d", actual.ID, expected.ID)
	}

	if actual.Title != expected.Title {
		t.Errorf(
			"unexpected title: got %q, want %q",
			actual.Title,
			expected.Title,
		)
	}

	if actual.Description != expected.Description {
		t.Errorf(
			"unexpected description: got %q, want %q",
			actual.Description,
			expected.Description,
		)
	}

	if actual.Done != expected.Done {
		t.Errorf(
			"unexpected done: got %t, want %t",
			actual.Done,
			expected.Done,
		)
	}

	if !actual.CreatedAt.Equal(expected.CreatedAt) {
		t.Errorf(
			"unexpected created_at: got %s, want %s",
			actual.CreatedAt,
			expected.CreatedAt,
		)
	}

	if !actual.UpdatedAt.Equal(expected.UpdatedAt) {
		t.Errorf(
			"unexpected updated_at: got %s, want %s",
			actual.UpdatedAt,
			expected.UpdatedAt,
		)
	}
}

func TestTasksRepositoryGetTaskNotFound(t *testing.T) {
	const missingTaskID int64 = 9999

	repository := newTestRepository(t)
	ctx := newTestContext(t)

	actual, err := repository.GetTask(ctx, missingTaskID)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf(
			"unexpected error: got %v, want ErrNotFound",
			err,
		)
	}

	if actual != (core_domain.Task{}) {
		t.Errorf(
			"unexpected task: got %+v, want zero task",
			actual,
		)
	}
}
