package tasks_service

import (
	"context"
	"fmt"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
)

func (s *TaskService) GetTask(
	ctx context.Context,
	taskID int64,
) (core_domain.Task, error) {
	task, err := s.taskRepository.GetTask(ctx, taskID)
	if err != nil {
		return core_domain.Task{}, fmt.Errorf("get task: %w", err)
	}

	return task, nil
}
