package tasks_postgres_repository

import (
	"context"
	"fmt"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
)

func (r *TasksRepository) GetTasks(
	ctx context.Context,
) ([]core_domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sql := `
	SELECT id, title, description, done, created_at, updated_at
	FROM checklist.tasks
	ORDER BY id DESC;
	`

	rows, err := r.pool.Query(
		ctx,
		sql,
	)
	if err != nil {
		return nil, fmt.Errorf("select tasks: %w", err)
	}
	defer rows.Close()

	var taskModels []TaskModel
	for rows.Next() {
		var taskModel TaskModel

		err := rows.Scan(
			&taskModel.ID,
			&taskModel.Title,
			&taskModel.Description,
			&taskModel.Done,
			&taskModel.CreatedAt,
			&taskModel.UpdateAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan tasks: %w", err)
		}

		taskModels = append(taskModels, taskModel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("next rows: %w", err)
	}

	taskDomain := taskDomainFromModels(taskModels)

	return taskDomain, nil
}
