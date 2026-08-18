package tasks_service

import (
	"context"
	"fmt"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
)

func (s *TaskService) CreateTask(
	ctx context.Context,
	task core_domain.Task,
) (core_domain.Task, error) {
	if err := task.Validate(); err != nil {
		return core_domain.Task{}, fmt.Errorf("validate task domain: %w", err)
	}

	task, err := s.taskRepository.CreateTask(ctx, task)
	if err != nil {
		return core_domain.Task{}, fmt.Errorf("crate task: %w", err)
	}

	return task, nil
}
