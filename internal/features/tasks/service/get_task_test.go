package tasks_service

import (
	"context"
	"errors"
	"testing"
	"time"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
)

func TestTaskServiceGetTaskSuccess(t *testing.T) {
	now := time.Date(2026, time.December, 25, 0, 0, 0, 0, time.UTC)

	expectedTask := core_domain.NewTask(
		defaultTaskID,
		"Buy groceries",
		"Milk and bread",
		false,
		core_domain.DefaultTaskPriority,
		now,
		now,
		nil,
	)

	var repositoryTaskID int64

	repository := taskRepositoryStub{
		getTaskFunc: func(
			_ context.Context,
			taskID int64,
		) (core_domain.Task, error) {
			repositoryTaskID = taskID
			return expectedTask, nil
		},
	}

	service := NewTaskService(repository)

	actualTask, err := service.GetTask(
		context.Background(),
		defaultTaskID,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repositoryTaskID != defaultTaskID {
		t.Errorf(
			"expected repository task ID %d, got %d",
			defaultTaskID,
			repositoryTaskID,
		)
	}

	if actualTask != expectedTask {
		t.Errorf(
			"expected task %+v, got %+v",
			expectedTask,
			actualTask,
		)
	}
}

func TestTaskServiceGetTaskRepositoryError(t *testing.T) {
	repositoryError := errors.New("database unavailable")

	var repositoryTaskID int64

	repository := taskRepositoryStub{
		getTaskFunc: func(
			_ context.Context,
			taskID int64,
		) (core_domain.Task, error) {
			repositoryTaskID = taskID

			return core_domain.Task{}, repositoryError
		},
	}

	service := NewTaskService(repository)

	actualTask, err := service.GetTask(
		context.Background(),
		defaultTaskID,
	)
	if !errors.Is(err, repositoryError) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}

	if repositoryTaskID != defaultTaskID {
		t.Errorf(
			"expected repository task ID %d, got %d",
			defaultTaskID,
			repositoryTaskID,
		)
	}

	if actualTask != (core_domain.Task{}) {
		t.Errorf(
			"expected empty task, got %+v",
			actualTask,
		)
	}
}
