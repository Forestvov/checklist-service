package core_domain

import (
	"errors"
	"testing"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

func TestTaskPriorityValidate(t *testing.T) {
	tests := []struct {
		name        string
		priority    TaskPriority
		expectError bool
	}{
		{
			name:     "low",
			priority: TaskPriorityLow,
		},
		{
			name:     "medium",
			priority: TaskPriorityMedium,
		},
		{
			name:     "high",
			priority: TaskPriorityHigh,
		},
		{
			name:        "empty",
			priority:    TaskPriority(""),
			expectError: true,
		},
		{
			name:        "unsupported",
			priority:    TaskPriority("unsupported"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.priority.Validate()

			if tt.expectError {
				if !errors.Is(err, core_errors.ErrInvalidArgument) {
					t.Fatalf(
						"expected ErrInvalidArgument, got %v",
						err,
					)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
