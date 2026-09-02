package core_domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

func TestTaskValidate(t *testing.T) {
	tests := []struct {
		name        string
		task        Task
		expectError bool
	}{
		{
			name: "valid task without description",
			task: Task{Title: "Task", Priority: DefaultTaskPriority},
		},
		{
			name: "valid maximum lengths",
			task: Task{
				Title:       strings.Repeat("я", 255),
				Description: strings.Repeat("я", 5000),
				Priority:    TaskPriorityHigh,
			},
		},
		{
			name:        "title is too short after trimming",
			task:        Task{Title: "  ab  ", Priority: DefaultTaskPriority},
			expectError: true,
		},
		{
			name:        "title is too long",
			task:        Task{Title: strings.Repeat("я", 256), Priority: DefaultTaskPriority},
			expectError: true,
		},
		{
			name: "description is too long",
			task: Task{
				Title:       "Task",
				Description: strings.Repeat("я", 5001),
				Priority:    DefaultTaskPriority,
			},
			expectError: true,
		},
		{
			name: "unsupported priority",
			task: Task{
				Title:    "Task",
				Priority: TaskPriority("critical"),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if tt.expectError && !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
			if !tt.expectError && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestNewTaskUninitializedDefaults(t *testing.T) {
	task := NewTaskUninitialized("Task", nil, nil, nil)

	if task.Description != "" {
		t.Fatalf("expected empty description, got %q", task.Description)
	}
	if task.Priority != DefaultTaskPriority {
		t.Errorf("expected priority %q, got %q", DefaultTaskPriority, task.Priority)
	}
	if task.DueAt != nil {
		t.Errorf("expected no deadline, got %v", task.DueAt)
	}
}

func TestUpdateTaskValidate(t *testing.T) {
	dueAt := time.Date(2026, time.September, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		update      UpdateTask
		expectError bool
	}{
		{
			name: "valid partial title update",
			update: UpdateTask{
				Title: setNullable("Updated task"),
			},
		},
		{
			name: "valid empty description",
			update: UpdateTask{
				Description: setNullable(""),
			},
		},
		{
			name: "valid false done value",
			update: UpdateTask{
				Done: setNullable(false),
			},
		},
		{
			name: "valid priority",
			update: UpdateTask{
				Priority: setNullable(TaskPriorityHigh),
			},
		},
		{
			name: "valid deadline",
			update: UpdateTask{
				DueAt: setNullable(dueAt),
			},
		},
		{
			name: "valid deadline removal",
			update: UpdateTask{
				DueAt: nullNullable[time.Time](),
			},
		},
		{
			name:        "no fields provided",
			expectError: true,
		},
		{
			name: "title is null",
			update: UpdateTask{
				Title: nullNullable[string](),
			},
			expectError: true,
		},
		{
			name: "title is too short after trimming",
			update: UpdateTask{
				Title: setNullable("  ab  "),
			},
			expectError: true,
		},
		{
			name: "title is too long",
			update: UpdateTask{
				Title: setNullable(strings.Repeat("я", 256)),
			},
			expectError: true,
		},
		{
			name: "description is null",
			update: UpdateTask{
				Description: nullNullable[string](),
			},
			expectError: true,
		},
		{
			name: "description is too long",
			update: UpdateTask{
				Description: setNullable(strings.Repeat("я", 5001)),
			},
			expectError: true,
		},
		{
			name: "done is null",
			update: UpdateTask{
				Done: nullNullable[bool](),
			},
			expectError: true,
		},
		{
			name: "priority is null",
			update: UpdateTask{
				Priority: nullNullable[TaskPriority](),
			},
			expectError: true,
		},
		{
			name: "priority is unsupported",
			update: UpdateTask{
				Priority: setNullable(TaskPriority("critical")),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.update.Validate()
			if tt.expectError && !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
			if !tt.expectError && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestTaskApplyUpdate(t *testing.T) {
	createdAt := time.Now().Add(-2 * time.Hour)
	updatedAt := time.Now().Add(-time.Hour)
	dueAt := time.Now().Add(24 * time.Hour)
	task := NewTask(
		1,
		"Original task",
		"Original description",
		false,
		DefaultTaskPriority,
		createdAt,
		updatedAt,
		nil,
	)

	patch := NewUpdateTask(
		setNullable("Updated task"),
		setNullable(""),
		setNullable(true),
		setNullable(TaskPriorityHigh),
		setNullable(dueAt),
	)

	if err := task.ApplyUpdate(patch); err != nil {
		t.Fatalf("apply task update: %v", err)
	}

	if task.Title != "Updated task" {
		t.Errorf("unexpected title: got %q, want %q", task.Title, "Updated task")
	}
	if task.Description != "" {
		t.Errorf("unexpected description: got %q, want empty", task.Description)
	}
	if !task.Done {
		t.Error("updated task must have done=true")
	}
	if task.Priority != TaskPriorityHigh {
		t.Errorf("unexpected priority: got %q, want %q", task.Priority, TaskPriorityHigh)
	}
	if task.DueAt == nil || !task.DueAt.Equal(dueAt) {
		t.Errorf("unexpected due_at: got %v, want %v", task.DueAt, dueAt)
	}
	if task.ID != 1 {
		t.Errorf("unexpected ID: got %d, want 1", task.ID)
	}
	if !task.CreatedAt.Equal(createdAt) {
		t.Errorf("created_at changed: got %s, want %s", task.CreatedAt, createdAt)
	}
	if !task.UpdatedAt.After(updatedAt) {
		t.Errorf(
			"updated_at was not refreshed: before %s, after %s",
			updatedAt,
			task.UpdatedAt,
		)
	}
}

func TestTaskApplyUpdateClearsDueAt(t *testing.T) {
	dueAt := time.Now().Add(24 * time.Hour)
	task := NewTask(
		1,
		"Original task",
		"",
		false,
		DefaultTaskPriority,
		time.Now().Add(-2*time.Hour),
		time.Now().Add(-time.Hour),
		&dueAt,
	)
	patch := NewUpdateTask(
		Nullable[string]{},
		Nullable[string]{},
		Nullable[bool]{},
		Nullable[TaskPriority]{},
		nullNullable[time.Time](),
	)

	if err := task.ApplyUpdate(patch); err != nil {
		t.Fatalf("clear task deadline: %v", err)
	}
	if task.DueAt != nil {
		t.Errorf("expected cleared due_at, got %v", task.DueAt)
	}
}

func TestTaskApplyPartialUpdate(t *testing.T) {
	updatedAt := time.Now().Add(-time.Hour)
	dueAt := time.Now().Add(24 * time.Hour)
	task := NewTask(
		1,
		"Original task",
		"Original description",
		false,
		DefaultTaskPriority,
		time.Now().Add(-2*time.Hour),
		updatedAt,
		&dueAt,
	)

	patch := NewUpdateTask(
		setNullable("Updated task"),
		Nullable[string]{},
		Nullable[bool]{},
		Nullable[TaskPriority]{},
		Nullable[time.Time]{},
	)

	if err := task.ApplyUpdate(patch); err != nil {
		t.Fatalf("apply partial task update: %v", err)
	}

	if task.Title != "Updated task" {
		t.Errorf("unexpected title: got %q, want %q", task.Title, "Updated task")
	}
	if task.Description != "Original description" {
		t.Errorf(
			"description changed: got %q, want %q",
			task.Description,
			"Original description",
		)
	}
	if task.Done {
		t.Error("done changed for omitted field")
	}
	if task.Priority != DefaultTaskPriority {
		t.Errorf("priority changed for omitted field: got %q", task.Priority)
	}
	if task.DueAt == nil || !task.DueAt.Equal(dueAt) {
		t.Errorf("due_at changed for omitted field: got %v, want %v", task.DueAt, dueAt)
	}
	if !task.UpdatedAt.After(updatedAt) {
		t.Errorf(
			"updated_at was not refreshed: before %s, after %s",
			updatedAt,
			task.UpdatedAt,
		)
	}
}

func TestTaskApplyInvalidUpdateDoesNotMutateTask(t *testing.T) {
	task := NewTask(
		1,
		"Original task",
		"Original description",
		false,
		DefaultTaskPriority,
		time.Now().Add(-2*time.Hour),
		time.Now().Add(-time.Hour),
		nil,
	)
	original := task

	patch := NewUpdateTask(
		setNullable("ab"),
		Nullable[string]{},
		Nullable[bool]{},
		Nullable[TaskPriority]{},
		Nullable[time.Time]{},
	)

	err := task.ApplyUpdate(patch)
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if task != original {
		t.Errorf("task changed after rejected update: got %+v, want %+v", task, original)
	}
}

func setNullable[T any](value T) Nullable[T] {
	return Nullable[T]{
		Value: &value,
		Set:   true,
	}
}

func nullNullable[T any]() Nullable[T] {
	return Nullable[T]{Set: true}
}
