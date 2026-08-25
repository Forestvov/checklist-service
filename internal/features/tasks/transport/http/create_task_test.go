package tasks_transport_http

import (
	"errors"
	"strings"
	"testing"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

func TestCreateTaskRequestValidate(t *testing.T) {
	tooLongDescription := strings.Repeat("я", 5001)

	tests := []struct {
		name        string
		request     CreateTaskRequest
		expectError bool
	}{
		{name: "valid request", request: CreateTaskRequest{Title: "Task"}},
		{
			name:        "title too short after trimming",
			request:     CreateTaskRequest{Title: "  a  "},
			expectError: true,
		},
		{
			name:        "description too long",
			request:     CreateTaskRequest{Title: "Task", Description: &tooLongDescription},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.expectError && !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
			if !tt.expectError && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
