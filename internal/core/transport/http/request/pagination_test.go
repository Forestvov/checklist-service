package core_http_request

import (
	"errors"
	"net/http/httptest"
	"testing"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
	core_pagination "github.com/Forestvov/checklist-service/internal/core/pagination"
)

func TestGetPaginationParams(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantPage    int
		wantPerPage int
		expectError bool
	}{
		{
			name:        "defaults",
			wantPage:    core_pagination.DefaultPage,
			wantPerPage: core_pagination.DefaultPerPage,
		},
		{name: "custom values", query: "?page=2&per_page=50", wantPage: 2, wantPerPage: 50},
		{name: "invalid page", query: "?page=task", expectError: true},
		{name: "zero page", query: "?page=0", expectError: true},
		{name: "per page over limit", query: "?per_page=101", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/tasks"+tt.query, nil)

			params, err := GetPaginationParams(r)
			if tt.expectError {
				if !errors.Is(err, core_errors.ErrInvalidArgument) {
					t.Fatalf("expected ErrInvalidArgument, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if params.Page != tt.wantPage {
				t.Errorf("expected page %d, got %d", tt.wantPage, params.Page)
			}
			if params.PerPage != tt.wantPerPage {
				t.Errorf("expected per_page %d, got %d", tt.wantPerPage, params.PerPage)
			}
		})
	}
}
