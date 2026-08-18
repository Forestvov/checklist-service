package tasks_service

import (
	"context"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
)

type TaskService struct {
	taskRepository TaskRepository
}

type TaskRepository interface {
	CreateTask(
		ctx context.Context,
		task core_domain.Task,
	) (core_domain.Task, error)
}

func NewTaskService(
	taskRepository TaskRepository,
) *TaskService {
	return &TaskService{
		taskRepository: taskRepository,
	}
}
