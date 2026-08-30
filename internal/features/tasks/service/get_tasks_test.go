package tasks_service

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
	core_pagination "github.com/Forestvov/checklist-service/internal/core/pagination"
)

func TestTaskServiceGetTasksSuccess(t *testing.T) {
	now := time.Date(2026, time.December, 25, 0, 0, 0, 0, time.UTC)

	paginationParams := core_pagination.Params{
		Page:    2,
		PerPage: 2,
	}

	expectedTasks := []core_domain.Task{
		core_domain.NewTask(
			10,
			"Buy groceries",
			"Milk and bread",
			false,
			now,
			now,
		),
		core_domain.NewTask(
			9,
			"Write tests",
			"Test service layer",
			true,
			now,
			now,
		),
	}

	expectedResult := core_pagination.NewResult(
		expectedTasks,
		5,
		paginationParams,
	)
	done := true
	filter := core_domain.NewTaskFilter(
		&done,
		core_domain.TaskSortTitle,
		core_domain.SortOrderAsc,
	)

	var (
		repositoryParams core_pagination.Params
		repositoryFilter core_domain.TaskFilter
	)

	repository := taskRepositoryStub{
		getTasksFunc: func(
			_ context.Context,
			params core_pagination.Params,
			filter core_domain.TaskFilter,
		) (core_pagination.Result[core_domain.Task], error) {
			repositoryParams = params
			repositoryFilter = filter

			return expectedResult, nil
		},
	}

	service := NewTaskService(repository)

	actualResult, err := service.GetTasks(
		context.Background(),
		paginationParams,
		filter,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repositoryParams != paginationParams {
		t.Errorf(
			"expected repository params %+v, got %+v",
			paginationParams,
			repositoryParams,
		)
	}
	if repositoryFilter.Done == nil || !*repositoryFilter.Done {
		t.Errorf("expected done=true filter, got %+v", repositoryFilter)
	}
	if repositoryFilter.Sort != core_domain.TaskSortTitle ||
		repositoryFilter.Order != core_domain.SortOrderAsc {
		t.Errorf("expected title asc sorting, got %+v", repositoryFilter)
	}

	if !slices.Equal(actualResult.Items, expectedResult.Items) {
		t.Errorf(
			"expected items %+v, got %+v",
			expectedResult.Items,
			actualResult.Items,
		)
	}

	if actualResult.Total != expectedResult.Total {
		t.Errorf(
			"expected total %d, got %d",
			expectedResult.Total,
			actualResult.Total,
		)
	}

	if actualResult.Params != expectedResult.Params {
		t.Errorf(
			"expected result params %+v, got %+v",
			expectedResult.Params,
			actualResult.Params,
		)
	}
}

func TestTaskServiceGetTasksInvalidFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter core_domain.TaskFilter
	}{
		{
			name: "unsupported sort",
			filter: core_domain.TaskFilter{
				Sort:  core_domain.TaskSort("priority"),
				Order: core_domain.SortOrderAsc,
			},
		},
		{
			name: "unsupported order",
			filter: core_domain.TaskFilter{
				Sort:  core_domain.TaskSortCreatedAt,
				Order: core_domain.SortOrder("sideways"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repositoryCalled := false
			repository := taskRepositoryStub{
				getTasksFunc: func(
					_ context.Context,
					_ core_pagination.Params,
					_ core_domain.TaskFilter,
				) (core_pagination.Result[core_domain.Task], error) {
					repositoryCalled = true
					return core_pagination.Result[core_domain.Task]{}, nil
				},
			}

			actual, err := NewTaskService(repository).GetTasks(
				context.Background(),
				core_pagination.Params{Page: 1, PerPage: 20},
				tt.filter,
			)
			if !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
			if repositoryCalled {
				t.Fatal("repository must not be called for an invalid filter")
			}
			if actual.Items != nil || actual.Total != 0 ||
				actual.Params != (core_pagination.Params{}) {
				t.Errorf("expected empty result, got %+v", actual)
			}
		})
	}
}

func TestTaskServiceGetTasksRepositoryError(t *testing.T) {
	paginationParams := core_pagination.Params{
		Page:    2,
		PerPage: 10,
	}

	repositoryError := errors.New("database unavailable")
	done := false
	filter := core_domain.NewTaskFilter(
		&done,
		"",
		"",
	)
	repositoryResult := core_pagination.NewResult(
		[]core_domain.Task{
			{ID: defaultTaskID, Title: "Unexpected task"},
		},
		1,
		paginationParams,
	)

	var (
		repositoryParams core_pagination.Params
		repositoryFilter core_domain.TaskFilter
	)

	repository := taskRepositoryStub{
		getTasksFunc: func(
			_ context.Context,
			params core_pagination.Params,
			filter core_domain.TaskFilter,
		) (core_pagination.Result[core_domain.Task], error) {
			repositoryParams = params
			repositoryFilter = filter

			return repositoryResult, repositoryError
		},
	}

	service := NewTaskService(repository)

	actualResult, err := service.GetTasks(
		context.Background(),
		paginationParams,
		filter,
	)

	if !errors.Is(err, repositoryError) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}

	if repositoryParams != paginationParams {
		t.Errorf(
			"expected repository params %+v, got %+v",
			paginationParams,
			repositoryParams,
		)
	}
	if repositoryFilter.Done == nil || *repositoryFilter.Done {
		t.Errorf("expected done=false filter, got %+v", repositoryFilter)
	}

	if actualResult.Items != nil {
		t.Errorf(
			"expected nil items, got %+v",
			actualResult.Items,
		)
	}

	if actualResult.Total != 0 {
		t.Errorf(
			"expected total 0, got %d",
			actualResult.Total,
		)
	}

	if actualResult.Params != (core_pagination.Params{}) {
		t.Errorf(
			"expected empty params, got %+v",
			actualResult.Params,
		)
	}
}
