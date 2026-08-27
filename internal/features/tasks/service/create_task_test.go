package tasks_service

import (
	"context"
	"errors"
	"testing"
	"time"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

func TestTaskServiceCreateTaskSuccess(t *testing.T) {
	description := "Milk and bread"
	inputTask := core_domain.NewTaskUninitialized(
		"Buy groceries",
		&description,
	)

	now := time.Date(2026, time.December, 25, 0, 0, 0, 0, time.UTC)
	expectedTask := core_domain.NewTask(
		42,
		inputTask.Title,
		inputTask.Description,
		false,
		now,
		now,
	)

	var repositoryArgument core_domain.Task

	repository := taskRepositoryStub{
		createTaskFunc: func(
			ctx context.Context,
			task core_domain.Task,
		) (core_domain.Task, error) {
			repositoryArgument = task
			return expectedTask, nil
		},
	}

	service := NewTaskService(repository)

	actualTask, err := service.CreateTask(context.Background(), inputTask)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repositoryArgument != inputTask {
		t.Errorf(
			"expected repository argument %+v, got %+v",
			inputTask,
			repositoryArgument,
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

func TestTaskServiceCreateTaskInvalidTask(t *testing.T) {
	inputTask := core_domain.NewTaskUninitialized(
		"a",
		nil,
	)

	repositoryCalled := false
	repository := taskRepositoryStub{
		createTaskFunc: func(
			_ context.Context,
			_ core_domain.Task,
		) (core_domain.Task, error) {
			repositoryCalled = true
			return core_domain.Task{}, nil
		},
	}

	service := NewTaskService(repository)

	actualTask, err := service.CreateTask(
		context.Background(),
		inputTask,
	)
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf(
			"expected ErrInvalidArgument, got %v",
			err,
		)
	}

	if repositoryCalled {
		t.Error("repository must not be called for an invalid task")
	}

	if actualTask != (core_domain.Task{}) {
		t.Errorf(
			"expected empty task, got %+v",
			actualTask,
		)
	}
}

func TestTaskServiceCreateTaskRepositoryError(t *testing.T) {
	inputTask := core_domain.NewTaskUninitialized(
		"Buy groceries",
		nil,
	)

	repositoryError := errors.New("database unavailable")
	repositoryCalled := false

	repository := taskRepositoryStub{
		createTaskFunc: func(
			_ context.Context,
			_ core_domain.Task,
		) (core_domain.Task, error) {
			repositoryCalled = true

			return core_domain.Task{}, repositoryError
		},
	}

	service := NewTaskService(repository)

	actualTask, err := service.CreateTask(
		context.Background(),
		inputTask,
	)

	if !errors.Is(err, repositoryError) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}

	if !repositoryCalled {
		t.Error("repository must be called for a valid task")
	}

	if actualTask != (core_domain.Task{}) {
		t.Errorf(
			"expected empty task, got %+v",
			actualTask,
		)
	}
}
