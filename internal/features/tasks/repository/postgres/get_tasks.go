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
	filter core_domain.TaskFilter,
) (core_pagination.Result[core_domain.Task], error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var emptyResult core_pagination.Result[core_domain.Task]

	countSQL := `
		SELECT COUNT(*)
		FROM checklist.tasks
		WHERE ($1::boolean IS NULL OR done = $1)
			AND ($2::varchar IS NULL OR priority = $2);
	`

	var done any
	if filter.Done != nil {
		done = *filter.Done
	}

	var priority any
	if filter.Priority != nil {
		priority = *filter.Priority
	}

	var total int64
	if err := r.pool.QueryRow(
		ctx,
		countSQL,
		done,
		priority,
	).Scan(&total); err != nil {
		return emptyResult, fmt.Errorf("count tasks: %w", err)
	}

	var orderColumn string
	switch filter.Sort {
	case core_domain.TaskSortCreatedAt:
		orderColumn = "created_at"
	case core_domain.TaskSortUpdatedAt:
		orderColumn = "updated_at"
	case core_domain.TaskSortTitle:
		orderColumn = "title"
	}

	var orderDirection string
	switch filter.Order {
	case core_domain.SortOrderAsc:
		orderDirection = "ASC"
	case core_domain.SortOrderDesc:
		orderDirection = "DESC"
	}

	orderBy := orderColumn + " " + orderDirection

	selectSQL := fmt.Sprintf(`
		SELECT id, title, description, done, priority, created_at, updated_at
		FROM checklist.tasks
		WHERE ($1::boolean IS NULL OR done = $1)
			AND ($2::varchar IS NULL OR priority = $2)
		ORDER BY %s, id %s
		LIMIT $3 OFFSET $4;
	`, orderBy, orderDirection)

	rows, err := r.pool.Query(
		ctx,
		selectSQL,
		done,
		priority,
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
			&taskModel.Priority,
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
