package tasks_postgres_repository

import (
	"context"
	"fmt"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

func (r *TasksRepository) DeleteTask(
	ctx context.Context,
	taskID int64,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sql := `
		DELETE FROM checklist.tasks WHERE id=$1;
	`

	cmdTag, err := r.pool.Exec(ctx, sql, taskID)
	if err != nil {
		return fmt.Errorf("exec query: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("task with id='%d': %w", taskID, core_errors.ErrNotFound)
	}

	return nil
}
