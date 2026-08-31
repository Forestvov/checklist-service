package tasks_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
	core_postgres_pool "github.com/Forestvov/checklist-service/internal/core/repository/postgres/pool"
)

func (r *TasksRepository) GetTask(
	ctx context.Context,
	taskID int64,
) (core_domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sql := `
		SELECT id, title, description, done, priority, created_at, updated_at
		FROM checklist.tasks
		WHERE id=$1;
	`

	row := r.pool.QueryRow(ctx, sql, taskID)

	var taskModel TaskModel
	err := row.Scan(
		&taskModel.ID,
		&taskModel.Title,
		&taskModel.Description,
		&taskModel.Done,
		&taskModel.Priority,
		&taskModel.CreatedAt,
		&taskModel.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return core_domain.Task{}, fmt.Errorf(
				"task with id=%d: %w",
				taskID,
				core_errors.ErrNotFound,
			)
		}

		return core_domain.Task{}, fmt.Errorf("scan task: %w", err)
	}

	taskDomain := taskDomainFromModel(taskModel)
	return taskDomain, nil
}
