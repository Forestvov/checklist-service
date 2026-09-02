package tasks_service

import (
	"context"
	"errors"
	"testing"
	"time"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

func TestTaskServiceUpdateTaskSuccess(t *testing.T) {
	createdAt := time.Date(2026, time.April, 14, 12, 0, 0, 0, time.UTC)
	dueAt := createdAt.Add(48 * time.Hour)
	original := core_domain.NewTask(
		defaultTaskID,
		"Buy groceries",
		"Milk and bread",
		false,
		core_domain.DefaultTaskPriority,
		createdAt,
		createdAt,
		nil,
	)

	newDescription := "Milk, bread and eggs"
	done := true
	patch := core_domain.NewUpdateTask(
		core_domain.Nullable[string]{},
		setDomainNullable(newDescription),
		setDomainNullable(done),
		setDomainNullable(core_domain.TaskPriorityHigh),
		setDomainNullable(dueAt),
	)

	repositoryResult := original
	repositoryResult.Description = newDescription
	repositoryResult.Done = done
	repositoryResult.Priority = core_domain.TaskPriorityHigh
	repositoryResult.DueAt = &dueAt
	repositoryResult.UpdatedAt = createdAt.Add(time.Hour)

	var (
		getTaskID       int64
		updateTaskID    int64
		updateTaskValue core_domain.Task
	)

	repository := taskRepositoryStub{
		getTaskFunc: func(_ context.Context, taskID int64) (core_domain.Task, error) {
			getTaskID = taskID
			return original, nil
		},
		updateTaskFunc: func(
			_ context.Context,
			taskID int64,
			task core_domain.Task,
		) (core_domain.Task, error) {
			updateTaskID = taskID
			updateTaskValue = task
			return repositoryResult, nil
		},
	}

	beforeUpdate := time.Now()
	actual, err := NewTaskService(repository).UpdateTask(
		context.Background(),
		defaultTaskID,
		patch,
	)
	afterUpdate := time.Now()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if getTaskID != defaultTaskID || updateTaskID != defaultTaskID {
		t.Errorf(
			"expected task ID %d, got get=%d update=%d",
			defaultTaskID,
			getTaskID,
			updateTaskID,
		)
	}
	if updateTaskValue.Title != original.Title {
		t.Errorf("expected unchanged title %q, got %q", original.Title, updateTaskValue.Title)
	}
	if updateTaskValue.Description != newDescription {
		t.Errorf("expected description %q, got %q", newDescription, updateTaskValue.Description)
	}
	if !updateTaskValue.Done {
		t.Error("expected done to be true")
	}
	if updateTaskValue.Priority != core_domain.TaskPriorityHigh {
		t.Errorf("expected priority %q, got %q", core_domain.TaskPriorityHigh, updateTaskValue.Priority)
	}
	if updateTaskValue.DueAt == nil || !updateTaskValue.DueAt.Equal(dueAt) {
		t.Errorf("expected due_at %v, got %v", dueAt, updateTaskValue.DueAt)
	}
	if updateTaskValue.UpdatedAt.Before(beforeUpdate) || updateTaskValue.UpdatedAt.After(afterUpdate) {
		t.Errorf("expected updated_at between %v and %v, got %v", beforeUpdate, afterUpdate, updateTaskValue.UpdatedAt)
	}
	if actual != repositoryResult {
		t.Errorf("expected task %+v, got %+v", repositoryResult, actual)
	}
}

func TestTaskServiceUpdateTaskGetTaskError(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	updateCalled := false
	repository := taskRepositoryStub{
		getTaskFunc: func(_ context.Context, _ int64) (core_domain.Task, error) {
			return core_domain.Task{}, repositoryError
		},
		updateTaskFunc: func(
			_ context.Context,
			_ int64,
			_ core_domain.Task,
		) (core_domain.Task, error) {
			updateCalled = true
			return core_domain.Task{}, nil
		},
	}

	actual, err := NewTaskService(repository).UpdateTask(
		context.Background(),
		defaultTaskID,
		core_domain.UpdateTask{},
	)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("expected repository error, got %v", err)
	}
	if updateCalled {
		t.Fatal("update must not be called when getting the task fails")
	}
	if actual != (core_domain.Task{}) {
		t.Errorf("expected empty task, got %+v", actual)
	}
}

func TestTaskServiceUpdateTaskInvalidPatch(t *testing.T) {
	original := core_domain.NewTask(
		defaultTaskID,
		"Buy groceries",
		"Milk and bread",
		false,
		core_domain.DefaultTaskPriority,
		time.Now(),
		time.Now(),
		nil,
	)

	updateCalled := false
	repository := taskRepositoryStub{
		getTaskFunc: func(_ context.Context, _ int64) (core_domain.Task, error) {
			return original, nil
		},
		updateTaskFunc: func(
			_ context.Context,
			_ int64,
			_ core_domain.Task,
		) (core_domain.Task, error) {
			updateCalled = true
			return core_domain.Task{}, nil
		},
	}

	actual, err := NewTaskService(repository).UpdateTask(
		context.Background(),
		defaultTaskID,
		core_domain.UpdateTask{},
	)
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected invalid argument error, got %v", err)
	}
	if updateCalled {
		t.Fatal("repository update must not be called for an invalid patch")
	}
	if actual != (core_domain.Task{}) {
		t.Errorf("expected empty task, got %+v", actual)
	}
}

func TestTaskServiceUpdateTaskRepositoryError(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	original := core_domain.NewTask(
		defaultTaskID,
		"Buy groceries",
		"Milk and bread",
		false,
		core_domain.DefaultTaskPriority,
		time.Now(),
		time.Now(),
		nil,
	)

	title := "Updated title"
	repository := taskRepositoryStub{
		getTaskFunc: func(_ context.Context, _ int64) (core_domain.Task, error) {
			return original, nil
		},
		updateTaskFunc: func(
			_ context.Context,
			_ int64,
			_ core_domain.Task,
		) (core_domain.Task, error) {
			return core_domain.Task{}, repositoryError
		},
	}

	actual, err := NewTaskService(repository).UpdateTask(
		context.Background(),
		defaultTaskID,
		core_domain.NewUpdateTask(
			setDomainNullable(title),
			core_domain.Nullable[string]{},
			core_domain.Nullable[bool]{},
			core_domain.Nullable[core_domain.TaskPriority]{},
			core_domain.Nullable[time.Time]{},
		),
	)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("expected repository error, got %v", err)
	}
	if actual != (core_domain.Task{}) {
		t.Errorf("expected empty task, got %+v", actual)
	}
}

func setDomainNullable[T any](value T) core_domain.Nullable[T] {
	return core_domain.Nullable[T]{Value: &value, Set: true}
}
