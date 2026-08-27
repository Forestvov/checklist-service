package tasks_service

import (
	"context"
	"errors"
	"testing"
	"time"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
)

func TestTaskServiceCompleteTaskSuccess(t *testing.T) {
	createdAt := time.Date(2026, time.April, 14, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)

	expectedTask := core_domain.NewTask(
		defaultTaskID,
		"Buy groceries",
		"Milk and bread",
		true,
		createdAt,
		updatedAt,
	)

	var repositoryTaskID int64

	repository := taskRepositoryStub{
		completeTaskFunc: func(
			_ context.Context,
			taskID int64,
		) (core_domain.Task, error) {
			repositoryTaskID = taskID

			return expectedTask, nil
		},
	}

	service := NewTaskService(repository)

	actualTask, err := service.CompleteTask(
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

func TestTaskServiceCompleteTaskRepositoryError(t *testing.T) {
	repositoryError := errors.New("database unavailable")

	var repositoryTaskID int64

	unexpectedTask := core_domain.Task{
		ID:    defaultTaskID,
		Title: "Unexpected task",
	}

	repository := taskRepositoryStub{
		completeTaskFunc: func(
			_ context.Context,
			taskID int64,
		) (core_domain.Task, error) {
			repositoryTaskID = taskID

			return unexpectedTask, repositoryError
		},
	}

	service := NewTaskService(repository)

	actualTask, err := service.CompleteTask(
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
