package tasks_service

import (
	"context"
	"fmt"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_pagination "github.com/Forestvov/checklist-service/internal/core/pagination"
)

func (s *TaskService) GetTasks(
	ctx context.Context,
	paginationParams core_pagination.Params,
) (core_pagination.Result[core_domain.Task], error) {
	result, err := s.taskRepository.GetTasks(
		ctx,
		paginationParams,
	)
	if err != nil {
		return core_pagination.Result[core_domain.Task]{}, fmt.Errorf(
			"get tasks from repository: %w",
			err,
		)
	}

	return result, nil
}
