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
	core_http_request "github.com/Forestvov/checklist-service/internal/core/transport/http/request"
)

func TestTasksHTTPHandlerCreateTaskSuccess(t *testing.T) {
	const expectedVersion int64 = 1

	now := time.Date(2026, time.December, 25, 0, 0, 0, 0, time.UTC)
	dueAt := time.Date(2027, time.January, 10, 12, 0, 0, 0, time.UTC)

	var serviceArgument core_domain.Task
	service := tasksServiceStub{
		createTaskFunc: func(
			ctx context.Context,
			task core_domain.Task,
		) (core_domain.Task, error) {
			serviceArgument = task

			created := core_domain.NewTask(
				defaultTaskID,
				task.Title,
				task.Description,
				false,
				task.Priority,
				now,
				now,
				task.DueAt,
				expectedVersion,
			)
			return created, nil
		},
	}

	handler := NewTasksHTTPHandler(service)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/tasks",
		strings.NewReader(`{
			"title": "Buy groceries",
			"description": "Milk and bread",
			"priority": "high",
			"due_at": "2027-01-10T12:00:00Z"
		}`),
	)
	request = requestWithTestLogger(request)

	recorder := httptest.NewRecorder()

	handler.CreateTask(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			recorder.Code,
		)
	}

	if serviceArgument.Title != "Buy groceries" {
		t.Errorf(
			"expected service title %q, got %q",
			"Buy groceries",
			serviceArgument.Title,
		)
	}
	if serviceArgument.Description != "Milk and bread" {
		t.Errorf(
			"expected service description %q, got %q",
			"Milk and bread",
			serviceArgument.Description,
		)
	}
	if serviceArgument.ID != core_domain.UninitializedID {
		t.Errorf(
			"expected uninitialized service task ID %d, got %d",
			core_domain.UninitializedID,
			serviceArgument.ID,
		)
	}
	if serviceArgument.Version != core_domain.UninitializedVersion {
		t.Errorf(
			"expected uninitialized service task version %d, got %d",
			core_domain.UninitializedVersion,
			serviceArgument.Version,
		)
	}
	if serviceArgument.Done {
		t.Error("expected uncompleted service task")
	}
	if serviceArgument.Priority != core_domain.TaskPriorityHigh {
		t.Errorf(
			"expected priority %q, got %q",
			core_domain.TaskPriorityHigh,
			serviceArgument.Priority,
		)
	}
	if serviceArgument.DueAt == nil || !serviceArgument.DueAt.Equal(dueAt) {
		t.Errorf("expected due_at %v, got %v", dueAt, serviceArgument.DueAt)
	}

	var response CreateTaskResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.ID != defaultTaskID {
		t.Errorf("expected task ID %d, got %d", defaultTaskID, response.ID)
	}
	if response.Version != expectedVersion {
		t.Errorf("expected response version %d, got %d", expectedVersion, response.Version)
	}

	if response.Title != "Buy groceries" {
		t.Errorf(
			"expected response title %q, got %q",
			"Buy groceries",
			response.Title,
		)
	}

	if response.Priority != core_domain.TaskPriorityHigh {
		t.Errorf(
			"expected response priority %q, got %q",
			core_domain.TaskPriorityHigh,
			response.Priority,
		)
	}
	if response.DueAt == nil || !response.DueAt.Equal(dueAt) {
		t.Errorf("expected response due_at %v, got %v", dueAt, response.DueAt)
	}
}

func TestTasksHTTPHandlerCreateTaskDefaultPriority(t *testing.T) {
	var serviceArgument core_domain.Task
	service := tasksServiceStub{
		createTaskFunc: func(
			_ context.Context,
			task core_domain.Task,
		) (core_domain.Task, error) {
			serviceArgument = task
			return task, nil
		},
	}
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/tasks",
		strings.NewReader(`{"title":"Buy groceries"}`),
	)
	request = requestWithTestLogger(request)
	recorder := httptest.NewRecorder()

	NewTasksHTTPHandler(service).CreateTask(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, recorder.Code)
	}
	if serviceArgument.Priority != core_domain.DefaultTaskPriority {
		t.Errorf(
			"expected default priority %q, got %q",
			core_domain.DefaultTaskPriority,
			serviceArgument.Priority,
		)
	}
	if serviceArgument.DueAt != nil {
		t.Errorf("expected no default deadline, got %v", serviceArgument.DueAt)
	}
}

func TestTasksHTTPHandlerCreateTaskInvalidRequest(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "empty body", body: ""},
		{name: "malformed JSON", body: `{"title":`},
		{name: "wrong priority type", body: `{"title":"Task","priority":1}`},
		{name: "unsupported priority", body: `{"title":"Task","priority":"critical"}`},
		{name: "invalid due_at", body: `{"title":"Task","due_at":"tomorrow"}`},
		{name: "unknown field", body: `{"title":"Task","unknown":true}`},
		{name: "multiple JSON values", body: `{"title":"Task"} {"title":"Other"}`},
		{name: "title too short after trimming", body: `{"title":"  a  "}`},
		{
			name: "body too large",
			body: strings.Repeat("a", int(core_http_request.MaxRequestBodySize)+1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceCalled := false
			service := tasksServiceStub{
				createTaskFunc: func(
					_ context.Context,
					_ core_domain.Task,
				) (core_domain.Task, error) {
					serviceCalled = true
					return core_domain.Task{}, nil
				},
			}

			handler := NewTasksHTTPHandler(service)
			request := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/tasks",
				strings.NewReader(tt.body),
			)
			request = requestWithTestLogger(request)
			recorder := httptest.NewRecorder()

			handler.CreateTask(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf(
					"expected status %d, got %d",
					http.StatusBadRequest,
					recorder.Code,
				)
			}
			if serviceCalled {
				t.Fatal("service must not be called for an invalid request")
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

func TestTasksHTTPHandlerCreateTaskServiceError(t *testing.T) {
	service := tasksServiceStub{
		createTaskFunc: func(
			_ context.Context,
			_ core_domain.Task,
		) (core_domain.Task, error) {
			return core_domain.Task{}, errors.New(internalServiceErrorMessage)
		},
	}

	handler := NewTasksHTTPHandler(service)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/tasks",
		strings.NewReader(`{"title":"Buy groceries"}`),
	)

	request = requestWithTestLogger(request)
	recorder := httptest.NewRecorder()

	handler.CreateTask(recorder, request)

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

func TestCreateTaskRequestValidate(t *testing.T) {
	tooLongDescription := strings.Repeat("я", 5001)
	highPriority := core_domain.TaskPriorityHigh
	unsupportedPriority := core_domain.TaskPriority("critical")

	tests := []struct {
		name        string
		request     CreateTaskRequest
		expectError bool
	}{
		{name: "valid request", request: CreateTaskRequest{Title: "Task"}},
		{
			name: "valid high priority",
			request: CreateTaskRequest{
				Title:    "Task",
				Priority: &highPriority,
			},
		},
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
		{
			name: "unsupported priority",
			request: CreateTaskRequest{
				Title:    "Task",
				Priority: &unsupportedPriority,
			},
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
