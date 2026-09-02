package tasks_postgres_repository

import (
	"context"
	"fmt"
	"time"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_pagination "github.com/Forestvov/checklist-service/internal/core/pagination"
)

func (r *TasksRepository) GetTasks(
	ctx context.Context,
	paginationParams core_pagination.Params,
	filter core_domain.TaskFilter,
	referenceTime time.Time,
) (core_pagination.Result[core_domain.Task], error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var emptyResult core_pagination.Result[core_domain.Task]

	var done any
	if filter.Done != nil {
		done = *filter.Done
	}

	var priority any
	if filter.Priority != nil {
		priority = *filter.Priority
	}

	var overdue any
	if filter.Overdue != nil {
		overdue = *filter.Overdue
	}

	countSQL := `
		SELECT COUNT(*)
		FROM checklist.tasks
		WHERE ($1::boolean IS NULL OR done = $1)
			AND ($2::varchar IS NULL OR priority = $2)
			AND (
				$3::boolean IS NULL
				OR (
					done = FALSE
					AND due_at IS NOT NULL
					AND due_at < $4::timestamptz
				) = $3
			);
	`

	var total int64
	if err := r.pool.QueryRow(
		ctx,
		countSQL,
		done,
		priority,
		overdue,
		referenceTime,
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
	case core_domain.TaskSortPriority:
		orderColumn = `CASE priority
			WHEN 'low' THEN 1
			WHEN 'medium' THEN 2
			WHEN 'high' THEN 3
		END`
	case core_domain.TaskSortDueAt:
		orderColumn = "due_at"
	}

	var orderDirection string
	switch filter.Order {
	case core_domain.SortOrderAsc:
		orderDirection = "ASC"
	case core_domain.SortOrderDesc:
		orderDirection = "DESC"
	}

	orderBy := orderColumn + " " + orderDirection
	if filter.Sort == core_domain.TaskSortDueAt {
		orderBy += " NULLS LAST"
	}

	selectSQL := fmt.Sprintf(`
		SELECT id, title, description, done, priority, created_at, updated_at, due_at, version
		FROM checklist.tasks
		WHERE ($1::boolean IS NULL OR done = $1)
			AND ($2::varchar IS NULL OR priority = $2)
			AND (
				$3::boolean IS NULL
				OR (
					done = FALSE
					AND due_at IS NOT NULL
					AND due_at < $4::timestamptz
				) = $3
			)
		ORDER BY %s, id %s
		LIMIT $5 OFFSET $6;
	`, orderBy, orderDirection)

	rows, err := r.pool.Query(
		ctx,
		selectSQL,
		done,
		priority,
		overdue,
		referenceTime,
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
			&taskModel.DueAt,
			&taskModel.Version,
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
