package tasks_service

import (
	"context"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_pagination "github.com/Forestvov/checklist-service/internal/core/pagination"
)

const defaultTaskID int64 = 10

type taskRepositoryStub struct {
	createTaskFunc func(
		ctx context.Context,
		task core_domain.Task,
	) (core_domain.Task, error)

	getTasksFunc func(
		ctx context.Context,
		paginationParams core_pagination.Params,
	) (core_pagination.Result[core_domain.Task], error)

	getTaskFunc func(
		ctx context.Context,
		taskID int64,
	) (core_domain.Task, error)

	completeTaskFunc func(
		ctx context.Context,
		taskID int64,
	) (core_domain.Task, error)

	deleteTaskFunc func(
		ctx context.Context,
		taskID int64,
	) error
}

func (s taskRepositoryStub) CreateTask(
	ctx context.Context,
	task core_domain.Task,
) (core_domain.Task, error) {
	return s.createTaskFunc(ctx, task)
}

func (s taskRepositoryStub) GetTasks(
	ctx context.Context,
	paginationParams core_pagination.Params,
) (core_pagination.Result[core_domain.Task], error) {
	return s.getTasksFunc(ctx, paginationParams)
}

func (s taskRepositoryStub) GetTask(
	ctx context.Context,
	taskID int64,
) (core_domain.Task, error) {
	return s.getTaskFunc(ctx, taskID)
}

func (s taskRepositoryStub) CompleteTask(
	ctx context.Context,
	taskID int64,
) (core_domain.Task, error) {
	return s.completeTaskFunc(ctx, taskID)
}

func (s taskRepositoryStub) DeleteTask(
	ctx context.Context,
	taskID int64,
) error {
	return s.deleteTaskFunc(ctx, taskID)
}
