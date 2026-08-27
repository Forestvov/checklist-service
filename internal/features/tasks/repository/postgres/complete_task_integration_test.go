//go:build integration

package tasks_postgres_repository

import (
	"errors"
	"testing"
	"time"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

func TestTasksRepositoryCompleteTaskSuccess(t *testing.T) {
	repository := newTestRepository(t)
	ctx := newTestContext(t)

	initialTime := time.Now().
		Add(-time.Hour).
		UTC().
		Truncate(time.Microsecond)
	input := core_domain.NewTask(
		core_domain.UninitializedID,
		"Buy groceries",
		"Milk, bread and vegetables",
		false,
		initialTime,
		initialTime,
	)

	created, err := repository.CreateTask(ctx, input)
	if err != nil {
		t.Fatalf("prepare task: %v", err)
	}

	if created.Done {
		t.Fatal("prepared task must not be completed")
	}

	completed, err := repository.CompleteTask(ctx, created.ID)
	if err != nil {
		t.Fatalf("complete task: %v", err)
	}

	if !completed.Done {
		t.Error("completed task must have done=true")
	}

	if completed.ID != created.ID {
		t.Errorf(
			"unexpected ID: got %d, want %d",
			completed.ID,
			created.ID,
		)
	}

	if completed.Title != created.Title {
		t.Errorf(
			"unexpected title: got %q, want %q",
			completed.Title,
			created.Title,
		)
	}

	if completed.Description != created.Description {
		t.Errorf(
			"unexpected description: got %q, want %q",
			completed.Description,
			created.Description,
		)
	}

	if !completed.CreatedAt.Equal(created.CreatedAt) {
		t.Errorf(
			"created_at changed: got %s, want %s",
			completed.CreatedAt,
			created.CreatedAt,
		)
	}

	if !completed.UpdatedAt.After(created.UpdatedAt) {
		t.Errorf(
			"updated_at was not refreshed: before %s, after %s",
			created.UpdatedAt,
			completed.UpdatedAt,
		)
	}

	stored, err := repository.GetTask(ctx, created.ID)
	if err != nil {
		t.Fatalf("get completed task: %v", err)
	}

	if !stored.Done {
		t.Error("stored task must have done=true")
	}

	if !stored.UpdatedAt.Equal(completed.UpdatedAt) {
		t.Errorf(
			"unexpected stored updated_at: got %s, want %s",
			stored.UpdatedAt,
			completed.UpdatedAt,
		)
	}
}

func TestTasksRepositoryCompleteTaskNotFound(t *testing.T) {
	const missingTaskID int64 = 9999

	repository := newTestRepository(t)
	ctx := newTestContext(t)

	actual, err := repository.CompleteTask(ctx, missingTaskID)
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
