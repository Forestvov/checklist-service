package tasks_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
	core_postgres_pool "github.com/Forestvov/checklist-service/internal/core/repository/postgres/pool"
)

func (r *TasksRepository) CompleteTask(
	ctx context.Context,
	taskID int64,
) (core_domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sql := `
		UPDATE checklist.tasks
		SET done = TRUE, updated_at = NOW()
		WHERE id = $1
		RETURNING id, title, description, done, created_at, updated_at;
	`

	row := r.pool.QueryRow(ctx, sql, taskID)

	var taskModel TaskModel
	err := row.Scan(
		&taskModel.ID,
		&taskModel.Title,
		&taskModel.Description,
		&taskModel.Done,
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

		return core_domain.Task{}, fmt.Errorf("scan completed task: %w", err)
	}

	return taskDomainFromModel(taskModel), nil
}
