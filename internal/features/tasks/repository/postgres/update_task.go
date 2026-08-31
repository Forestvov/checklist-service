package tasks_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
	core_postgres_pool "github.com/Forestvov/checklist-service/internal/core/repository/postgres/pool"
)

func (r *TasksRepository) UpdateTask(
	ctx context.Context,
	taskID int64,
	taskUpdate core_domain.Task,
) (core_domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sql := `
		UPDATE checklist.tasks
		SET title = $2,
			description = $3,
			done = $4,
			priority = $5,
			updated_at = $6
		WHERE id = $1
		RETURNING id, title, description, done, priority, created_at, updated_at;
		`

	row := r.pool.QueryRow(
		ctx,
		sql,
		taskID,
		taskUpdate.Title,
		taskUpdate.Description,
		taskUpdate.Done,
		taskUpdate.Priority,
		taskUpdate.UpdatedAt,
	)

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

		return core_domain.Task{}, fmt.Errorf("scan updated task: %w", err)
	}

	return taskDomainFromModel(taskModel), nil
}
