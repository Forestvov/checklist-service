package tasks_service

import (
	"context"
	"fmt"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
)

func (s *TaskService) UpdateTask(
	ctx context.Context,
	taskID int64,
	taskUpdate core_domain.UpdateTask,
) (core_domain.Task, error) {
	task, err := s.taskRepository.GetTask(ctx, taskID)
	if err != nil {
		return core_domain.Task{}, fmt.Errorf("get task: %w", err)
	}

	if err := task.ApplyUpdate(taskUpdate); err != nil {
		return core_domain.Task{}, fmt.Errorf("apply task patch: %w", err)
	}

	patchedTask, err := s.taskRepository.UpdateTask(ctx, taskID, task)
	if err != nil {
		return core_domain.Task{}, fmt.Errorf("patch task: %w", err)
	}

	return patchedTask, nil
}
