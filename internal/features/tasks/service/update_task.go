package tasks_service

import (
	"context"
	"fmt"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

func (s *TaskService) UpdateTask(
	ctx context.Context,
	taskID int64,
	expectedVersion int64,
	taskUpdate core_domain.UpdateTask,
) (core_domain.Task, error) {
	if expectedVersion <= 0 {
		return core_domain.Task{}, fmt.Errorf(
			"expected version must be greater than zero: %w",
			core_errors.ErrInvalidArgument,
		)
	}

	if err := taskUpdate.Validate(); err != nil {
		return core_domain.Task{}, fmt.Errorf(
			"task update validation failed: %w",
			err,
		)
	}

	updatedTask, err := s.taskRepository.UpdateTask(
		ctx,
		taskID,
		expectedVersion,
		taskUpdate,
	)
	if err != nil {
		return core_domain.Task{}, fmt.Errorf(
			"update task: %w",
			err,
		)
	}

	return updatedTask, nil
}
