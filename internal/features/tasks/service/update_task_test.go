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
	const (
		expectedVersion int64 = 3
		resultVersion   int64 = 4
	)

	title := "Updated task"
	patch := core_domain.NewUpdateTask(
		setDomainNullable(title),
		core_domain.Nullable[string]{},
		core_domain.Nullable[bool]{},
		core_domain.Nullable[core_domain.TaskPriority]{},
		core_domain.Nullable[time.Time]{},
	)

	now := time.Date(2026, time.April, 14, 12, 0, 0, 0, time.UTC)
	expectedTask := core_domain.NewTask(
		defaultTaskID,
		title,
		"Milk and bread",
		false,
		core_domain.DefaultTaskPriority,
		now,
		now.Add(time.Hour),
		nil,
		resultVersion,
	)

	var (
		repositoryTaskID          int64
		repositoryExpectedVersion int64
		repositoryPatch           core_domain.UpdateTask
	)

	repository := taskRepositoryStub{
		updateTaskFunc: func(
			_ context.Context,
			taskID int64,
			version int64,
			taskUpdate core_domain.UpdateTask,
		) (core_domain.Task, error) {
			repositoryTaskID = taskID
			repositoryExpectedVersion = version
			repositoryPatch = taskUpdate
			return expectedTask, nil
		},
	}

	actual, err := NewTaskService(repository).UpdateTask(
		context.Background(),
		defaultTaskID,
		expectedVersion,
		patch,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repositoryTaskID != defaultTaskID {
		t.Errorf("expected repository task ID %d, got %d", defaultTaskID, repositoryTaskID)
	}
	if repositoryExpectedVersion != expectedVersion {
		t.Errorf(
			"expected repository version %d, got %d",
			expectedVersion,
			repositoryExpectedVersion,
		)
	}
	if !repositoryPatch.Title.Set ||
		repositoryPatch.Title.Value == nil ||
		*repositoryPatch.Title.Value != title {
		t.Errorf("unexpected repository patch: %+v", repositoryPatch)
	}
	if actual != expectedTask {
		t.Errorf("expected task %+v, got %+v", expectedTask, actual)
	}
}

func TestTaskServiceUpdateTaskInvalidVersion(t *testing.T) {
	title := "Updated task"
	patch := core_domain.NewUpdateTask(
		setDomainNullable(title),
		core_domain.Nullable[string]{},
		core_domain.Nullable[bool]{},
		core_domain.Nullable[core_domain.TaskPriority]{},
		core_domain.Nullable[time.Time]{},
	)

	tests := []struct {
		name    string
		version int64
	}{
		{name: "zero", version: 0},
		{name: "negative", version: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repositoryCalled := false
			repository := taskRepositoryStub{
				updateTaskFunc: func(
					_ context.Context,
					_ int64,
					_ int64,
					_ core_domain.UpdateTask,
				) (core_domain.Task, error) {
					repositoryCalled = true
					return core_domain.Task{}, nil
				},
			}

			actual, err := NewTaskService(repository).UpdateTask(
				context.Background(),
				defaultTaskID,
				tt.version,
				patch,
			)
			if !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
			if repositoryCalled {
				t.Fatal("repository must not be called for an invalid version")
			}
			if actual != (core_domain.Task{}) {
				t.Errorf("expected empty task, got %+v", actual)
			}
		})
	}
}

func TestTaskServiceUpdateTaskInvalidPatch(t *testing.T) {
	repositoryCalled := false
	repository := taskRepositoryStub{
		updateTaskFunc: func(
			_ context.Context,
			_ int64,
			_ int64,
			_ core_domain.UpdateTask,
		) (core_domain.Task, error) {
			repositoryCalled = true
			return core_domain.Task{}, nil
		},
	}

	actual, err := NewTaskService(repository).UpdateTask(
		context.Background(),
		defaultTaskID,
		3,
		core_domain.UpdateTask{},
	)
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if repositoryCalled {
		t.Fatal("repository must not be called for an invalid patch")
	}
	if actual != (core_domain.Task{}) {
		t.Errorf("expected empty task, got %+v", actual)
	}
}

func TestTaskServiceUpdateTaskRepositoryError(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	title := "Updated title"
	patch := core_domain.NewUpdateTask(
		setDomainNullable(title),
		core_domain.Nullable[string]{},
		core_domain.Nullable[bool]{},
		core_domain.Nullable[core_domain.TaskPriority]{},
		core_domain.Nullable[time.Time]{},
	)

	repository := taskRepositoryStub{
		updateTaskFunc: func(
			_ context.Context,
			_ int64,
			_ int64,
			_ core_domain.UpdateTask,
		) (core_domain.Task, error) {
			return core_domain.Task{}, repositoryError
		},
	}

	actual, err := NewTaskService(repository).UpdateTask(
		context.Background(),
		defaultTaskID,
		3,
		patch,
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
