package core_domain

import (
	"errors"
	"testing"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

func TestNewTaskFilter(t *testing.T) {
	done := true
	priority := TaskPriorityHigh

	tests := []struct {
		name      string
		done      *bool
		priority  *TaskPriority
		sort      TaskSort
		order     SortOrder
		wantSort  TaskSort
		wantOrder SortOrder
	}{
		{
			name:      "defaults",
			wantSort:  DefaultTaskSort,
			wantOrder: DefaultSortOrder,
		},
		{
			name:      "provided values",
			done:      &done,
			priority:  &priority,
			sort:      TaskSortTitle,
			order:     SortOrderAsc,
			wantSort:  TaskSortTitle,
			wantOrder: SortOrderAsc,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewTaskFilter(tt.done, tt.priority, tt.sort, tt.order)

			if filter.Done != tt.done {
				t.Errorf("expected done pointer %p, got %p", tt.done, filter.Done)
			}
			if filter.Priority != tt.priority {
				t.Errorf("expected priority pointer %p, got %p", tt.priority, filter.Priority)
			}
			if filter.Sort != tt.wantSort {
				t.Errorf("expected sort %q, got %q", tt.wantSort, filter.Sort)
			}
			if filter.Order != tt.wantOrder {
				t.Errorf("expected order %q, got %q", tt.wantOrder, filter.Order)
			}
		})
	}
}

func TestTaskFilterValidate(t *testing.T) {
	validPriority := TaskPriorityHigh
	invalidPriority := TaskPriority("critical")

	tests := []struct {
		name        string
		filter      TaskFilter
		expectError bool
	}{
		{
			name:   "created at ascending",
			filter: NewTaskFilter(nil, nil, TaskSortCreatedAt, SortOrderAsc),
		},
		{
			name:   "updated at descending",
			filter: NewTaskFilter(nil, nil, TaskSortUpdatedAt, SortOrderDesc),
		},
		{
			name:   "title ascending",
			filter: NewTaskFilter(nil, nil, TaskSortTitle, SortOrderAsc),
		},
		{
			name:   "priority descending",
			filter: NewTaskFilter(nil, nil, TaskSortPriority, SortOrderDesc),
		},
		{
			name:   "due at ascending",
			filter: NewTaskFilter(nil, nil, TaskSortDueAt, SortOrderAsc),
		},
		{
			name:   "priority filter",
			filter: NewTaskFilter(nil, &validPriority, TaskSortCreatedAt, SortOrderDesc),
		},
		{
			name: "unsupported sort",
			filter: TaskFilter{
				Sort:  TaskSort("deadline"),
				Order: SortOrderAsc,
			},
			expectError: true,
		},
		{
			name: "unsupported order",
			filter: TaskFilter{
				Sort:  TaskSortCreatedAt,
				Order: SortOrder("sideways"),
			},
			expectError: true,
		},
		{
			name: "unsupported priority",
			filter: TaskFilter{
				Priority: &invalidPriority,
				Sort:     TaskSortCreatedAt,
				Order:    SortOrderDesc,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()
			if tt.expectError {
				if !errors.Is(err, core_errors.ErrInvalidArgument) {
					t.Fatalf("expected ErrInvalidArgument, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
