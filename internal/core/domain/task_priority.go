package core_domain

import (
	"fmt"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"

	DefaultTaskPriority = TaskPriorityMedium
)

func (p TaskPriority) Validate() error {
	switch p {
	case TaskPriorityLow, TaskPriorityMedium, TaskPriorityHigh:
		return nil
	default:
		return fmt.Errorf(
			"unsupported task priority %q: %w",
			p,
			core_errors.ErrInvalidArgument,
		)
	}
}
