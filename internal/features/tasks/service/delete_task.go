package tasks_service

import (
	"context"
	"fmt"
)

func (s *TaskService) DeleteTask(
	ctx context.Context,
	taskID int64,
) error {
	if err := s.taskRepository.DeleteTask(
		ctx,
		taskID,
	); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	return nil
}
