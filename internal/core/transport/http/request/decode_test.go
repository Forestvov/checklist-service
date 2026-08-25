package core_http_request

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

type decodeTestRequest struct {
	Name string `json:"name" validate:"required"`
}

func TestDecodeValidateRequest(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantName    string
		expectError bool
	}{
		{name: "valid request", body: `{"name":"task"}`, wantName: "task"},
		{name: "unknown field", body: `{"name":"task","unexpected":true}`, expectError: true},
		{name: "empty body", body: "  \n", expectError: true},
		{name: "malformed JSON", body: `{"name":`, expectError: true},
		{name: "multiple JSON values", body: `{"name":"task"} {"name":"other"}`, expectError: true},
		{name: "validation failure", body: `{"name":""}`, expectError: true},
		{
			name:        "body too large",
			body:        strings.Repeat("a", int(MaxRequestBodySize)+1),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/", strings.NewReader(tt.body))
			var request decodeTestRequest

			err := DecodeValidateRequest(r, &request)
			if tt.expectError {
				if !errors.Is(err, core_errors.ErrInvalidArgument) {
					t.Fatalf("expected ErrInvalidArgument, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if request.Name != tt.wantName {
				t.Errorf("expected name %q, got %q", tt.wantName, request.Name)
			}
		})
	}
}
