package core_pagination

import "testing"

func TestResultTotalPages(t *testing.T) {
	tests := []struct {
		name    string
		total   int64
		perPage int
		want    int64
	}{
		{name: "empty result", total: 0, perPage: 20, want: 0},
		{name: "complete pages", total: 40, perPage: 20, want: 2},
		{name: "partial last page", total: 47, perPage: 20, want: 3},
		{name: "invalid per page", total: 47, perPage: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewResult(
				[]int{},
				tt.total,
				Params{Page: 1, PerPage: tt.perPage},
			)

			if got := result.TotalPages(); got != tt.want {
				t.Fatalf("expected %d pages, got %d", tt.want, got)
			}
		})
	}
}

func TestNewResultNormalizesNilItems(t *testing.T) {
	result := NewResult[int](nil, 0, Params{Page: 1, PerPage: 20})

	if result.Items == nil {
		t.Fatal("expected an empty slice, got nil")
	}
}
