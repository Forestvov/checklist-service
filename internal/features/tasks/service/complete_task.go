package tasks_service

import (
	"context"
	"fmt"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
)

func (s *TaskService) CompleteTask(
	ctx context.Context,
	taskID int64,
) (core_domain.Task, error) {
	task, err := s.taskRepository.CompleteTask(ctx, taskID)
	if err != nil {
		return core_domain.Task{}, fmt.Errorf("complete task: %w", err)
	}

	return task, nil
}
