package tasks_service

import (
	"context"
	"errors"
	"testing"
)

func TestTaskServiceDeleteTaskSuccess(t *testing.T) {
	var repositoryTaskID int64

	repository := taskRepositoryStub{
		deleteTaskFunc: func(
			_ context.Context,
			taskID int64,
		) error {
			repositoryTaskID = taskID

			return nil
		},
	}
	service := NewTaskService(repository)

	err := service.DeleteTask(context.Background(), defaultTaskID)
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
}

func TestTaskServiceDeleteTaskRepositoryError(t *testing.T) {
	repositoryError := errors.New("database unavailable")

	var repositoryTaskID int64

	repository := taskRepositoryStub{
		deleteTaskFunc: func(
			_ context.Context,
			taskID int64,
		) error {
			repositoryTaskID = taskID

			return repositoryError
		},
	}

	service := NewTaskService(repository)

	err := service.DeleteTask(
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
}
