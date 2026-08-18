package core_domain

import (
	"errors"
	"strings"
	"testing"

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
			task: Task{Title: "Task"},
		},
		{
			name: "valid maximum lengths",
			task: Task{
				Title:       strings.Repeat("я", 255),
				Description: strings.Repeat("я", 5000),
			},
		},
		{
			name:        "title is too short after trimming",
			task:        Task{Title: "  ab  "},
			expectError: true,
		},
		{
			name:        "title is too long",
			task:        Task{Title: strings.Repeat("я", 256)},
			expectError: true,
		},
		{
			name: "description is too long",
			task: Task{
				Title:       "Task",
				Description: strings.Repeat("я", 5001),
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

func TestNewTaskUninitializedWithoutDescription(t *testing.T) {
	task := NewTaskUninitialized("Task", nil)

	if task.Description != "" {
		t.Fatalf("expected empty description, got %q", task.Description)
	}
}
