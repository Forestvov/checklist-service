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
	expectedVersion int64,
	taskUpdate core_domain.UpdateTask,
) (core_domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sql := `
		UPDATE checklist.tasks
		SET title = CASE
				WHEN $3 THEN $4::varchar
				ELSE title
			END,
			description = CASE
				WHEN $5 THEN $6::text
				ELSE description
			END,
			done = CASE
				WHEN $7 THEN $8::boolean
				ELSE done
			END,
			priority = CASE
				WHEN $9 THEN $10::varchar
				ELSE priority
			END,
			due_at = CASE
				WHEN $11 THEN $12::timestamptz
				ELSE due_at
			END,
			updated_at = NOW(),
			version = version + 1
		WHERE id = $1
			AND version = $2
		RETURNING
			id,
			title,
			description,
			done,
			priority,
			created_at,
			updated_at,
			due_at,
			version;
	`

	row := r.pool.QueryRow(
		ctx,
		sql,
		taskID,
		expectedVersion,

		taskUpdate.Title.Set,
		taskUpdate.Title.Value,

		taskUpdate.Description.Set,
		taskUpdate.Description.Value,

		taskUpdate.Done.Set,
		taskUpdate.Done.Value,

		taskUpdate.Priority.Set,
		taskUpdate.Priority.Value,

		taskUpdate.DueAt.Set,
		taskUpdate.DueAt.Value,
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
		&taskModel.DueAt,
		&taskModel.Version,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return core_domain.Task{}, r.classifyUpdateMiss(
				ctx,
				taskID,
				expectedVersion,
			)
		}

		return core_domain.Task{}, fmt.Errorf("scan updated task: %w", err)
	}

	return taskDomainFromModel(taskModel), nil
}

func (r *TasksRepository) classifyUpdateMiss(
	ctx context.Context,
	taskID int64,
	expectedVersion int64,
) error {
	const sql = `
	SELECT version
	FROM checklist.tasks
	WHERE id = $1;
	`

	var actualVersion int64
	err := r.pool.QueryRow(ctx, sql, taskID).Scan(&actualVersion)

	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return fmt.Errorf(
			"task with id=%d: %w",
			taskID,
			core_errors.ErrNotFound,
		)
	}

	if err != nil {
		return fmt.Errorf(
			"get current version for task with id=%d: %w",
			taskID,
			err,
		)
	}

	return fmt.Errorf(
		"task with id=%d has version=%d, expected=%d: %w",
		taskID,
		actualVersion,
		expectedVersion,
		core_errors.ErrConflict,
	)
}
