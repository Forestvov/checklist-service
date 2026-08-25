package core_http_request

import (
	"errors"
	"net/http/httptest"
	"testing"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

func TestGetInt64PathValue(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		want        int64
		expectError bool
	}{
		{name: "positive ID", value: "42", want: 42},
		{name: "missing ID", expectError: true},
		{name: "zero ID", value: "0", expectError: true},
		{name: "negative ID", value: "-1", expectError: true},
		{name: "invalid ID", value: "task", expectError: true},
		{name: "overflow", value: "9223372036854775808", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/tasks/"+tt.value, nil)
			if tt.value != "" {
				r.SetPathValue("id", tt.value)
			}

			got, err := GetInt64PathValue(r, "id")
			if tt.expectError {
				if !errors.Is(err, core_errors.ErrInvalidArgument) {
					t.Fatalf("expected ErrInvalidArgument, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if got != tt.want {
				t.Errorf("expected %d, got %d", tt.want, got)
			}
		})
	}
}
