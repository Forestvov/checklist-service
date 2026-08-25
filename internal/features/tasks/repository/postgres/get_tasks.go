package tasks_postgres_repository

import (
	"context"
	"fmt"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_pagination "github.com/Forestvov/checklist-service/internal/core/pagination"
)

func (r *TasksRepository) GetTasks(
	ctx context.Context,
	paginationParams core_pagination.Params,
) (core_pagination.Result[core_domain.Task], error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var emptyResult core_pagination.Result[core_domain.Task]

	countSQL := `
		SELECT COUNT(*)
		FROM checklist.tasks;
	`

	var total int64
	if err := r.pool.QueryRow(ctx, countSQL).Scan(&total); err != nil {
		return emptyResult, fmt.Errorf("count tasks: %w", err)
	}

	selectSQL := `
		SELECT id, title, description, done, created_at, updated_at
		FROM checklist.tasks
		ORDER BY id DESC
		LIMIT $1 OFFSET $2;
	`

	rows, err := r.pool.Query(
		ctx,
		selectSQL,
		paginationParams.Limit(),
		paginationParams.Offset(),
	)
	if err != nil {
		return emptyResult, fmt.Errorf("select tasks: %w", err)
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
			&taskModel.UpdatedAt,
		)
		if err != nil {
			return emptyResult, fmt.Errorf("scan tasks: %w", err)
		}

		taskModels = append(taskModels, taskModel)
	}
	if err := rows.Err(); err != nil {
		return emptyResult, fmt.Errorf("iterate task rows: %w", err)
	}

	taskDomains := taskDomainFromModels(taskModels)

	result := core_pagination.NewResult(
		taskDomains,
		total,
		paginationParams,
	)

	return result, nil
}
