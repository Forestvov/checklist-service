package tasks_service

import (
	"context"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_pagination "github.com/Forestvov/checklist-service/internal/core/pagination"
)

type TaskService struct {
	taskRepository TaskRepository
}

type TaskRepository interface {
	CreateTask(
		ctx context.Context,
		task core_domain.Task,
	) (core_domain.Task, error)

	GetTasks(
		ctx context.Context,
		paginationParams core_pagination.Params,
	) (core_pagination.Result[core_domain.Task], error)

	GetTask(
		ctx context.Context,
		taskID int64,
	) (core_domain.Task, error)

	CompleteTask(
		ctx context.Context,
		taskID int64,
	) (core_domain.Task, error)

	DeleteTask(
		ctx context.Context,
		taskID int64,
	) error
}

func NewTaskService(
	taskRepository TaskRepository,
) *TaskService {
	return &TaskService{
		taskRepository: taskRepository,
	}
}
