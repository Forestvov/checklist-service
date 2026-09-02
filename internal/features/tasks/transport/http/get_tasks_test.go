package tasks_transport_http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
	core_pagination "github.com/Forestvov/checklist-service/internal/core/pagination"
)

func TestTasksHTTPHandlerGetTasksSuccess(t *testing.T) {
	now := time.Date(
		2026,
		time.December,
		25,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	dueAt := now.Add(24 * time.Hour)

	expectedTasks := []core_domain.Task{
		core_domain.NewTask(
			42,
			"Buy groceries",
			"Milk and bread",
			false,
			core_domain.TaskPriorityHigh,
			now,
			now,
			&dueAt,
			3,
		),

		core_domain.NewTask(
			41,
			"Write tests",
			"Test GetTasks handler",
			true,
			core_domain.TaskPriorityLow,
			now,
			now,
			nil,
			7,
		),
	}

	const totalTasks int64 = 5

	var (
		serviceParams core_pagination.Params
		serviceFilter core_domain.TaskFilter
	)

	service := tasksServiceStub{
		getTasksFunc: func(
			_ context.Context,
			paginationParams core_pagination.Params,
			filter core_domain.TaskFilter,
		) (core_pagination.Result[core_domain.Task], error) {
			serviceParams = paginationParams
			serviceFilter = filter

			return core_pagination.NewResult(
				expectedTasks,
				totalTasks,
				paginationParams,
			), nil
		},
	}

	handler := NewTasksHTTPHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/tasks?page=2&per_page=2",
		nil,
	)
	request = requestWithTestLogger(request)

	recorder := httptest.NewRecorder()

	handler.GetTasks(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	if serviceParams.Page != 2 {
		t.Errorf(
			"expected service page %d, got %d",
			2,
			serviceParams.Page,
		)
	}

	if serviceParams.PerPage != 2 {
		t.Errorf(
			"expected service per_page %d, got %d",
			2,
			serviceParams.PerPage,
		)
	}
	if serviceFilter.Done != nil {
		t.Errorf("expected empty done filter, got %+v", serviceFilter)
	}
	if serviceFilter.Priority != nil {
		t.Errorf("expected empty priority filter, got %+v", serviceFilter)
	}
	if serviceFilter.Overdue != nil {
		t.Errorf("expected empty overdue filter, got %+v", serviceFilter)
	}
	if serviceFilter.Sort != core_domain.DefaultTaskSort ||
		serviceFilter.Order != core_domain.DefaultSortOrder {
		t.Errorf("expected default sorting, got %+v", serviceFilter)
	}

	var response GetTasksResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(response.Data) != len(expectedTasks) {
		t.Fatalf(
			"expected %d tasks, got %d",
			len(expectedTasks),
			len(response.Data),
		)
	}

	if response.Data[0].ID != expectedTasks[0].ID {
		t.Errorf(
			"expected first task ID %d, got %d",
			expectedTasks[0].ID,
			response.Data[0].ID,
		)
	}
	if response.Data[0].Version != expectedTasks[0].Version {
		t.Errorf(
			"expected first task version %d, got %d",
			expectedTasks[0].Version,
			response.Data[0].Version,
		)
	}
	if response.Data[0].Priority != expectedTasks[0].Priority {
		t.Errorf(
			"expected first task priority %q, got %q",
			expectedTasks[0].Priority,
			response.Data[0].Priority,
		)
	}
	if response.Data[0].DueAt == nil || !response.Data[0].DueAt.Equal(dueAt) {
		t.Errorf("expected first task due_at %v, got %v", dueAt, response.Data[0].DueAt)
	}

	if response.Data[1].ID != expectedTasks[1].ID {
		t.Errorf(
			"expected second task ID %d, got %d",
			expectedTasks[1].ID,
			response.Data[1].ID,
		)
	}
	if response.Data[1].Version != expectedTasks[1].Version {
		t.Errorf(
			"expected second task version %d, got %d",
			expectedTasks[1].Version,
			response.Data[1].Version,
		)
	}
	if response.Data[1].Priority != expectedTasks[1].Priority {
		t.Errorf(
			"expected second task priority %q, got %q",
			expectedTasks[1].Priority,
			response.Data[1].Priority,
		)
	}
	if response.Data[1].DueAt != nil {
		t.Errorf("expected second task without due_at, got %v", response.Data[1].DueAt)
	}

	if response.Meta.Page != 2 {
		t.Errorf(
			"expected meta page %d, got %d",
			2,
			response.Meta.Page,
		)
	}

	if response.Meta.PerPage != 2 {
		t.Errorf(
			"expected meta per_page %d, got %d",
			2,
			response.Meta.PerPage,
		)
	}

	if response.Meta.Total != totalTasks {
		t.Errorf(
			"expected total %d, got %d",
			totalTasks,
			response.Meta.Total,
		)
	}

	if response.Meta.TotalPages != 3 {
		t.Errorf(
			"expected total pages %d, got %d",
			3,
			response.Meta.TotalPages,
		)
	}
}

func TestTasksHTTPHandlerGetTasksDefaultPagination(t *testing.T) {
	var serviceParams core_pagination.Params

	service := tasksServiceStub{
		getTasksFunc: func(
			_ context.Context,
			params core_pagination.Params,
			filter core_domain.TaskFilter,
		) (core_pagination.Result[core_domain.Task], error) {
			serviceParams = params

			return core_pagination.NewResult[core_domain.Task](
				nil,
				0,
				params,
			), nil
		},
	}

	handler := NewTasksHTTPHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/tasks",
		nil,
	)
	request = requestWithTestLogger(request)

	recorder := httptest.NewRecorder()

	handler.GetTasks(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	if serviceParams.Page != core_pagination.DefaultPage {
		t.Fatalf(
			"expected page %d, got %d",
			core_pagination.DefaultPage,
			serviceParams.Page,
		)
	}

	if serviceParams.PerPage != core_pagination.DefaultPerPage {
		t.Fatalf(
			"expected per_page %d, got %d",
			core_pagination.DefaultPerPage,
			serviceParams.PerPage,
		)
	}

	var response GetTasksResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Data == nil {
		t.Error("expected data to be an empty array, got null")
	}

	if len(response.Data) != 0 {
		t.Errorf(
			"expected 0 tasks, got %d",
			len(response.Data),
		)
	}

	if response.Meta.Page != core_pagination.DefaultPage {
		t.Errorf(
			"expected meta page %d, got %d",
			core_pagination.DefaultPage,
			response.Meta.Page,
		)
	}

	if response.Meta.PerPage != core_pagination.DefaultPerPage {
		t.Errorf(
			"expected meta per_page %d, got %d",
			core_pagination.DefaultPerPage,
			response.Meta.PerPage,
		)
	}

	if response.Meta.Total != 0 {
		t.Fatalf(
			"expected total %d, got %d",
			0,
			response.Meta.Total,
		)
	}

	if response.Meta.TotalPages != 0 {
		t.Errorf(
			"expected total pages %d, got %d",
			0,
			response.Meta.TotalPages,
		)
	}
}

func TestTasksHTTPHandlerGetTasksDoneFilter(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  bool
	}{
		{name: "completed tasks", query: "?done=true", want: true},
		{name: "uncompleted tasks", query: "?done=false", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serviceFilter core_domain.TaskFilter
			service := tasksServiceStub{
				getTasksFunc: func(
					_ context.Context,
					params core_pagination.Params,
					filter core_domain.TaskFilter,
				) (core_pagination.Result[core_domain.Task], error) {
					serviceFilter = filter
					return core_pagination.NewResult[core_domain.Task](nil, 0, params), nil
				},
			}
			request := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/tasks"+tt.query,
				nil,
			)
			request = requestWithTestLogger(request)
			recorder := httptest.NewRecorder()

			NewTasksHTTPHandler(service).GetTasks(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
			}
			if serviceFilter.Done == nil {
				t.Fatal("expected done filter, got nil")
			}
			if *serviceFilter.Done != tt.want {
				t.Errorf("expected done=%t, got %t", tt.want, *serviceFilter.Done)
			}
		})
	}
}

func TestTasksHTTPHandlerGetTasksOverdueFilter(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  bool
	}{
		{name: "overdue tasks", query: "?overdue=true", want: true},
		{name: "not overdue tasks", query: "?overdue=false", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serviceFilter core_domain.TaskFilter
			service := tasksServiceStub{
				getTasksFunc: func(
					_ context.Context,
					params core_pagination.Params,
					filter core_domain.TaskFilter,
				) (core_pagination.Result[core_domain.Task], error) {
					serviceFilter = filter
					return core_pagination.NewResult[core_domain.Task](nil, 0, params), nil
				},
			}
			request := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/tasks"+tt.query,
				nil,
			)
			request = requestWithTestLogger(request)
			recorder := httptest.NewRecorder()

			NewTasksHTTPHandler(service).GetTasks(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
			}
			if serviceFilter.Overdue == nil {
				t.Fatal("expected overdue filter, got nil")
			}
			if *serviceFilter.Overdue != tt.want {
				t.Errorf("expected overdue=%t, got %t", tt.want, *serviceFilter.Overdue)
			}
		})
	}
}

func TestTasksHTTPHandlerGetTasksPriorityFilter(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		want      core_domain.TaskPriority
		checkDone bool
		wantDone  bool
	}{
		{name: "low priority", query: "?priority=low", want: core_domain.TaskPriorityLow},
		{name: "medium priority", query: "?priority=medium", want: core_domain.TaskPriorityMedium},
		{name: "high priority", query: "?priority=high", want: core_domain.TaskPriorityHigh},
		{
			name:      "priority and done",
			query:     "?priority=high&done=false",
			want:      core_domain.TaskPriorityHigh,
			checkDone: true,
			wantDone:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serviceFilter core_domain.TaskFilter
			service := tasksServiceStub{
				getTasksFunc: func(
					_ context.Context,
					params core_pagination.Params,
					filter core_domain.TaskFilter,
				) (core_pagination.Result[core_domain.Task], error) {
					serviceFilter = filter
					return core_pagination.NewResult[core_domain.Task](nil, 0, params), nil
				},
			}
			request := httptest.NewRequest(http.MethodGet, "/api/v1/tasks"+tt.query, nil)
			request = requestWithTestLogger(request)
			recorder := httptest.NewRecorder()

			NewTasksHTTPHandler(service).GetTasks(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
			}
			if serviceFilter.Priority == nil {
				t.Fatal("expected priority filter, got nil")
			}
			if *serviceFilter.Priority != tt.want {
				t.Errorf("expected priority=%q, got %q", tt.want, *serviceFilter.Priority)
			}
			if tt.checkDone {
				if serviceFilter.Done == nil {
					t.Fatal("expected done filter, got nil")
				}
				if *serviceFilter.Done != tt.wantDone {
					t.Errorf("expected done=%t, got %t", tt.wantDone, *serviceFilter.Done)
				}
			}
		})
	}
}

func TestTasksHTTPHandlerGetTasksSorting(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantSort  core_domain.TaskSort
		wantOrder core_domain.SortOrder
	}{
		{
			name:      "default sorting",
			wantSort:  core_domain.DefaultTaskSort,
			wantOrder: core_domain.DefaultSortOrder,
		},
		{
			name:      "title ascending",
			query:     "?sort=title&order=asc",
			wantSort:  core_domain.TaskSortTitle,
			wantOrder: core_domain.SortOrderAsc,
		},
		{
			name:      "updated at descending",
			query:     "?sort=updated_at&order=desc",
			wantSort:  core_domain.TaskSortUpdatedAt,
			wantOrder: core_domain.SortOrderDesc,
		},
		{
			name:      "priority ascending",
			query:     "?sort=priority&order=asc",
			wantSort:  core_domain.TaskSortPriority,
			wantOrder: core_domain.SortOrderAsc,
		},
		{
			name:      "priority descending",
			query:     "?sort=priority&order=desc",
			wantSort:  core_domain.TaskSortPriority,
			wantOrder: core_domain.SortOrderDesc,
		},
		{
			name:      "due at ascending",
			query:     "?sort=due_at&order=asc",
			wantSort:  core_domain.TaskSortDueAt,
			wantOrder: core_domain.SortOrderAsc,
		},
		{
			name:      "due at descending",
			query:     "?sort=due_at&order=desc",
			wantSort:  core_domain.TaskSortDueAt,
			wantOrder: core_domain.SortOrderDesc,
		},
		{
			name:      "default order",
			query:     "?sort=title",
			wantSort:  core_domain.TaskSortTitle,
			wantOrder: core_domain.DefaultSortOrder,
		},
		{
			name:      "default sort field",
			query:     "?order=asc",
			wantSort:  core_domain.DefaultTaskSort,
			wantOrder: core_domain.SortOrderAsc,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serviceFilter core_domain.TaskFilter
			service := tasksServiceStub{
				getTasksFunc: func(
					_ context.Context,
					params core_pagination.Params,
					filter core_domain.TaskFilter,
				) (core_pagination.Result[core_domain.Task], error) {
					serviceFilter = filter
					return core_pagination.NewResult[core_domain.Task](nil, 0, params), nil
				},
			}
			request := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/tasks"+tt.query,
				nil,
			)
			request = requestWithTestLogger(request)
			recorder := httptest.NewRecorder()

			NewTasksHTTPHandler(service).GetTasks(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
			}
			if serviceFilter.Sort != tt.wantSort {
				t.Errorf("expected sort %q, got %q", tt.wantSort, serviceFilter.Sort)
			}
			if serviceFilter.Order != tt.wantOrder {
				t.Errorf("expected order %q, got %q", tt.wantOrder, serviceFilter.Order)
			}
		})
	}
}

func TestTasksHTTPHandlerGetTasksInvalidSorting(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "unsupported sort", query: "?sort=deadline"},
		{name: "unsupported order", query: "?order=sideways"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tasksServiceStub{
				getTasksFunc: func(
					_ context.Context,
					_ core_pagination.Params,
					filter core_domain.TaskFilter,
				) (core_pagination.Result[core_domain.Task], error) {
					return core_pagination.Result[core_domain.Task]{}, filter.Validate()
				},
			}
			request := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/tasks"+tt.query,
				nil,
			)
			request = requestWithTestLogger(request)
			recorder := httptest.NewRecorder()

			NewTasksHTTPHandler(service).GetTasks(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
			}
			response := decodeErrorResponse(t, recorder)
			if response.Error != core_errors.ErrInvalidArgument.Error() {
				t.Errorf(
					"expected error %q, got %q",
					core_errors.ErrInvalidArgument,
					response.Error,
				)
			}
		})
	}
}

func TestTasksHTTPHandlerGetTasksInvalidDoneFilter(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "empty value", query: "?done="},
		{name: "invalid value", query: "?done=task"},
		{name: "multiple values", query: "?done=true&done=false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceCalled := false
			service := tasksServiceStub{
				getTasksFunc: func(
					_ context.Context,
					_ core_pagination.Params,
					_ core_domain.TaskFilter,
				) (core_pagination.Result[core_domain.Task], error) {
					serviceCalled = true
					return core_pagination.Result[core_domain.Task]{}, nil
				},
			}
			request := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/tasks"+tt.query,
				nil,
			)
			request = requestWithTestLogger(request)
			recorder := httptest.NewRecorder()

			NewTasksHTTPHandler(service).GetTasks(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
			}
			if serviceCalled {
				t.Fatal("service must not be called for an invalid done filter")
			}
			response := decodeErrorResponse(t, recorder)
			if response.Error != core_errors.ErrInvalidArgument.Error() {
				t.Errorf(
					"expected error %q, got %q",
					core_errors.ErrInvalidArgument,
					response.Error,
				)
			}
		})
	}
}

func TestTasksHTTPHandlerGetTasksInvalidOverdueFilter(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "empty value", query: "?overdue="},
		{name: "invalid value", query: "?overdue=task"},
		{name: "multiple values", query: "?overdue=true&overdue=false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceCalled := false
			service := tasksServiceStub{
				getTasksFunc: func(
					_ context.Context,
					_ core_pagination.Params,
					_ core_domain.TaskFilter,
				) (core_pagination.Result[core_domain.Task], error) {
					serviceCalled = true
					return core_pagination.Result[core_domain.Task]{}, nil
				},
			}
			request := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/tasks"+tt.query,
				nil,
			)
			request = requestWithTestLogger(request)
			recorder := httptest.NewRecorder()

			NewTasksHTTPHandler(service).GetTasks(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
			}
			if serviceCalled {
				t.Fatal("service must not be called for an invalid overdue filter")
			}
			response := decodeErrorResponse(t, recorder)
			if response.Error != core_errors.ErrInvalidArgument.Error() {
				t.Errorf(
					"expected error %q, got %q",
					core_errors.ErrInvalidArgument,
					response.Error,
				)
			}
		})
	}
}

func TestTasksHTTPHandlerGetTasksInvalidPriorityFilter(t *testing.T) {
	tests := []struct {
		name              string
		query             string
		wantServiceCalled bool
	}{
		{name: "empty value", query: "?priority="},
		{name: "multiple values", query: "?priority=low&priority=high"},
		{name: "unsupported value", query: "?priority=critical", wantServiceCalled: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceCalled := false
			service := tasksServiceStub{
				getTasksFunc: func(
					_ context.Context,
					_ core_pagination.Params,
					filter core_domain.TaskFilter,
				) (core_pagination.Result[core_domain.Task], error) {
					serviceCalled = true
					return core_pagination.Result[core_domain.Task]{}, filter.Validate()
				},
			}
			request := httptest.NewRequest(http.MethodGet, "/api/v1/tasks"+tt.query, nil)
			request = requestWithTestLogger(request)
			recorder := httptest.NewRecorder()

			NewTasksHTTPHandler(service).GetTasks(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
			}
			if serviceCalled != tt.wantServiceCalled {
				t.Errorf("unexpected service call state: got %t, want %t", serviceCalled, tt.wantServiceCalled)
			}
			response := decodeErrorResponse(t, recorder)
			if response.Error != core_errors.ErrInvalidArgument.Error() {
				t.Errorf("expected error %q, got %q", core_errors.ErrInvalidArgument, response.Error)
			}
		})
	}
}

func TestTasksHTTPHandlerGetTasksInvalidPagination(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "zero page", query: "?page=0"},
		{name: "negative page", query: "?page=-1"},
		{name: "invalid page", query: "?page=task"},
		{name: "zero per page", query: "?per_page=0"},
		{name: "negative per page", query: "?per_page=-1"},
		{name: "per page over limit", query: "?per_page=101"},
		{name: "invalid per page", query: "?per_page=task"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceCalled := false
			service := tasksServiceStub{
				getTasksFunc: func(
					_ context.Context,
					_ core_pagination.Params,
					_ core_domain.TaskFilter,
				) (core_pagination.Result[core_domain.Task], error) {
					serviceCalled = true
					return core_pagination.Result[core_domain.Task]{}, nil
				},
			}

			handler := NewTasksHTTPHandler(service)
			request := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/tasks"+tt.query,
				nil,
			)

			request = requestWithTestLogger(request)
			recorder := httptest.NewRecorder()

			handler.GetTasks(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf(
					"expected status %d, got %d",
					http.StatusBadRequest,
					recorder.Code,
				)
			}
			if serviceCalled {
				t.Fatal("service must not be called for invalid pagination parameters")
			}

			response := decodeErrorResponse(t, recorder)
			if response.Error != core_errors.ErrInvalidArgument.Error() {
				t.Errorf(
					"expected error %q, got %q",
					core_errors.ErrInvalidArgument.Error(),
					response.Error,
				)
			}
		})
	}
}

func TestTasksHTTPHandlerGetTasksServiceError(t *testing.T) {
	service := tasksServiceStub{
		getTasksFunc: func(
			_ context.Context,
			params core_pagination.Params,
			_ core_domain.TaskFilter,
		) (core_pagination.Result[core_domain.Task], error) {
			return core_pagination.NewResult[core_domain.Task](
				nil,
				0,
				params,
			), errors.New(internalServiceErrorMessage)
		},
	}

	handler := NewTasksHTTPHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/tasks",
		nil,
	)
	request = requestWithTestLogger(request)

	recorder := httptest.NewRecorder()

	handler.GetTasks(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			recorder.Code,
		)
	}
	if strings.Contains(recorder.Body.String(), internalServiceErrorMessage) {
		t.Error("internal service error must not be exposed in the response")
	}

	response := decodeErrorResponse(t, recorder)
	if response.Error != http.StatusText(http.StatusInternalServerError) {
		t.Errorf(
			"expected error %q, got %q",
			http.StatusText(http.StatusInternalServerError),
			response.Error,
		)
	}
}
