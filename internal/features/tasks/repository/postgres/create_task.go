package tasks_postgres_repository

import (
	"context"
	"fmt"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
)

func (r *TasksRepository) CreateTask(
	ctx context.Context,
	task core_domain.Task,
) (core_domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sql := `
		INSERT INTO checklist.tasks (title, description, done, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, title, description, done, created_at, updated_at;
	`

	row := r.pool.QueryRow(
		ctx,
		sql,
		task.Title,
		task.Description,
		task.Done,
		task.CreatedAt,
		task.UpdatedAt,
	)

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
		return core_domain.Task{}, fmt.Errorf("scan created task: %w", err)
	}

	taskDomain := core_domain.NewTask(
		taskModel.ID,
		taskModel.Title,
		taskModel.Description,
		taskModel.Done,
		taskModel.CreatedAt,
		taskModel.UpdatedAt,
	)

	return taskDomain, nil
}
