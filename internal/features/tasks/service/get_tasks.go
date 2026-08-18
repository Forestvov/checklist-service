package tasks_service

import (
	"context"
	"fmt"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
)

func (s *TaskService) GetTasks(
	ctx context.Context,
) ([]core_domain.Task, error) {
	tasks, err := s.taskRepository.GetTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tasks from repository: %w", err)
	}

	return tasks, nil
}
