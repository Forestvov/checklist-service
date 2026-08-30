package core_domain

import (
	"fmt"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

type TaskSort string
type SortOrder string

const (
	TaskSortCreatedAt TaskSort = "created_at"
	TaskSortUpdatedAt TaskSort = "updated_at"
	TaskSortTitle     TaskSort = "title"

	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

type TaskFilter struct {
	Done  *bool
	Sort  TaskSort
	Order SortOrder
}

const (
	DefaultTaskSort  = TaskSortCreatedAt
	DefaultSortOrder = SortOrderDesc
)

func NewTaskFilter(
	done *bool,
	sortBy TaskSort,
	order SortOrder,
) TaskFilter {
	if sortBy == "" {
		sortBy = DefaultTaskSort
	}

	if order == "" {
		order = DefaultSortOrder
	}

	return TaskFilter{
		Done:  done,
		Sort:  sortBy,
		Order: order,
	}
}

func (f TaskFilter) Validate() error {
	switch f.Sort {
	case TaskSortCreatedAt, TaskSortUpdatedAt, TaskSortTitle:
	default:
		return fmt.Errorf(
			"unsupported task sort %q: %w",
			f.Sort,
			core_errors.ErrInvalidArgument,
		)
	}

	switch f.Order {
	case SortOrderAsc, SortOrderDesc:
	default:
		return fmt.Errorf(
			"unsupported sort order %q: %w",
			f.Order,
			core_errors.ErrInvalidArgument,
		)
	}

	return nil
}
