//go:build integration

package tasks_postgres_repository

import (
	"errors"
	"testing"
	"time"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

func TestTasksRepositoryUpdateTaskSuccess(t *testing.T) {
	repository := newTestRepository(t)
	ctx := newTestContext(t)

	created, err := repository.CreateTask(
		ctx,
		core_domain.NewTaskUninitialized("Original task", nil, nil),
	)
	if err != nil {
		t.Fatalf("prepare task: %v", err)
	}

	updated := created
	updated.Title = "Updated task"
	updated.Description = "Updated description"
	updated.Done = true
	updated.Priority = core_domain.TaskPriorityHigh
	updated.UpdatedAt = created.UpdatedAt.Add(time.Hour)

	actual, err := repository.UpdateTask(ctx, created.ID, updated)
	if err != nil {
		t.Fatalf("update task: %v", err)
	}

	assertTasksEqual(t, actual, updated)

	stored, err := repository.GetTask(ctx, created.ID)
	if err != nil {
		t.Fatalf("get updated task: %v", err)
	}

	assertTasksEqual(t, stored, updated)
}

func TestTasksRepositoryUpdateTaskNotFound(t *testing.T) {
	const missingTaskID int64 = 9999

	repository := newTestRepository(t)
	ctx := newTestContext(t)

	actual, err := repository.UpdateTask(
		ctx,
		missingTaskID,
		core_domain.Task{
			Title:       "Updated task",
			Description: "Updated description",
			Done:        true,
			Priority:    core_domain.TaskPriorityHigh,
			UpdatedAt:   time.Now(),
		},
	)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("unexpected error: got %v, want ErrNotFound", err)
	}
	if actual != (core_domain.Task{}) {
		t.Errorf("unexpected task: got %+v, want zero task", actual)
	}
}

func TestTasksRepositoryUpdateTaskRejectsInvalidTask(t *testing.T) {
	repository := newTestRepository(t)
	ctx := newTestContext(t)

	created, err := repository.CreateTask(
		ctx,
		core_domain.NewTaskUninitialized("Original task", nil, nil),
	)
	if err != nil {
		t.Fatalf("prepare task: %v", err)
	}

	invalid := created
	invalid.Title = "ab"
	invalid.UpdatedAt = created.UpdatedAt.Add(time.Hour)

	actual, err := repository.UpdateTask(ctx, created.ID, invalid)
	if err == nil {
		t.Fatal("expected database constraint error, got nil")
	}
	if actual != (core_domain.Task{}) {
		t.Errorf("unexpected task: got %+v, want zero task", actual)
	}

	stored, err := repository.GetTask(ctx, created.ID)
	if err != nil {
		t.Fatalf("get task after rejected update: %v", err)
	}

	assertTasksEqual(t, stored, created)
}

func assertTasksEqual(t *testing.T, actual, expected core_domain.Task) {
	t.Helper()

	if actual.ID != expected.ID {
		t.Errorf("unexpected ID: got %d, want %d", actual.ID, expected.ID)
	}
	if actual.Title != expected.Title {
		t.Errorf("unexpected title: got %q, want %q", actual.Title, expected.Title)
	}
	if actual.Description != expected.Description {
		t.Errorf(
			"unexpected description: got %q, want %q",
			actual.Description,
			expected.Description,
		)
	}
	if actual.Done != expected.Done {
		t.Errorf("unexpected done: got %t, want %t", actual.Done, expected.Done)
	}
	if actual.Priority != expected.Priority {
		t.Errorf("unexpected priority: got %q, want %q", actual.Priority, expected.Priority)
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
