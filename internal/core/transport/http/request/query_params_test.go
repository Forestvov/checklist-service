package core_http_request

import (
	"errors"
	"net/http/httptest"
	"testing"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

func TestGetBoolQueryParam(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		expectValue *bool
		expectError bool
	}{
		{name: "missing parameter"},
		{name: "true", query: "?done=true", expectValue: boolPointer(true)},
		{name: "false", query: "?done=false", expectValue: boolPointer(false)},
		{name: "empty value", query: "?done=", expectError: true},
		{name: "invalid value", query: "?done=task", expectError: true},
		{name: "multiple values", query: "?done=true&done=false", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/tasks"+tt.query, nil)

			actual, err := GetBoolQueryParam(request, "done")
			if tt.expectError {
				if !errors.Is(err, core_errors.ErrInvalidArgument) {
					t.Fatalf("expected ErrInvalidArgument, got %v", err)
				}
				if actual != nil {
					t.Errorf("expected nil value, got %t", *actual)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tt.expectValue == nil {
				if actual != nil {
					t.Errorf("expected nil value, got %t", *actual)
				}
				return
			}
			if actual == nil || *actual != *tt.expectValue {
				t.Errorf("expected %t, got %v", *tt.expectValue, actual)
			}
		})
	}
}

func boolPointer(value bool) *bool {
	return &value
}
