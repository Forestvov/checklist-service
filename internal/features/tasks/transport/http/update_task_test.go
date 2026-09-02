package tasks_transport_http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

func TestTasksHTTPHandlerUpdateTaskSuccess(t *testing.T) {
	const (
		requestVersion  int64 = 2
		responseVersion int64 = 3
	)

	createdAt := time.Date(2026, time.December, 25, 0, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	dueAt := createdAt.Add(24 * time.Hour)
	expectedTask := core_domain.NewTask(
		defaultTaskID,
		"Buy groceries",
		"Milk, bread and eggs",
		true,
		core_domain.TaskPriorityHigh,
		createdAt,
		updatedAt,
		&dueAt,
		responseVersion,
	)

	var (
		serviceTaskID          int64
		serviceExpectedVersion int64
		servicePatch           core_domain.UpdateTask
	)
	service := tasksServiceStub{
		updateTaskFunc: func(
			_ context.Context,
			taskID int64,
			expectedVersion int64,
			patch core_domain.UpdateTask,
		) (core_domain.Task, error) {
			serviceTaskID = taskID
			serviceExpectedVersion = expectedVersion
			servicePatch = patch
			return expectedTask, nil
		},
	}

	taskIDPath := strconv.FormatInt(defaultTaskID, 10)
	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/tasks/"+taskIDPath,
		strings.NewReader(`{"version":2,"description":"Milk, bread and eggs","done":true,"priority":"high","due_at":"2026-12-26T00:00:00Z"}`),
	)
	request.SetPathValue("id", taskIDPath)
	request = requestWithTestLogger(request)
	recorder := httptest.NewRecorder()

	NewTasksHTTPHandler(service).UpdateTask(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if serviceTaskID != defaultTaskID {
		t.Errorf("expected service task ID %d, got %d", defaultTaskID, serviceTaskID)
	}
	if serviceExpectedVersion != requestVersion {
		t.Errorf(
			"expected service version %d, got %d",
			requestVersion,
			serviceExpectedVersion,
		)
	}
	if servicePatch.Title.Set {
		t.Error("expected omitted title to remain unset")
	}
	if !servicePatch.Description.Set || servicePatch.Description.Value == nil || *servicePatch.Description.Value != expectedTask.Description {
		t.Errorf("unexpected description patch: %+v", servicePatch.Description)
	}
	if !servicePatch.Done.Set || servicePatch.Done.Value == nil || !*servicePatch.Done.Value {
		t.Errorf("unexpected done patch: %+v", servicePatch.Done)
	}
	if !servicePatch.Priority.Set ||
		servicePatch.Priority.Value == nil ||
		*servicePatch.Priority.Value != core_domain.TaskPriorityHigh {
		t.Errorf("unexpected priority patch: %+v", servicePatch.Priority)
	}
	if !servicePatch.DueAt.Set ||
		servicePatch.DueAt.Value == nil ||
		!servicePatch.DueAt.Value.Equal(dueAt) {
		t.Errorf("unexpected due_at patch: %+v", servicePatch.DueAt)
	}

	var response UpdateTaskResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ID != expectedTask.ID ||
		response.Title != expectedTask.Title ||
		response.Description != expectedTask.Description ||
		response.Done != expectedTask.Done ||
		response.Priority != expectedTask.Priority ||
		response.DueAt == nil ||
		!response.DueAt.Equal(*expectedTask.DueAt) ||
		!response.CreatedAt.Equal(expectedTask.CreatedAt) ||
		!response.UpdatedAt.Equal(expectedTask.UpdatedAt) ||
		response.Version != expectedTask.Version {
		t.Errorf("expected response for task %+v, got %+v", expectedTask, response)
	}
}

func TestTasksHTTPHandlerUpdateTaskClearDueAt(t *testing.T) {
	const (
		requestVersion  int64 = 2
		responseVersion int64 = 3
	)

	var (
		serviceExpectedVersion int64
		servicePatch           core_domain.UpdateTask
	)
	service := tasksServiceStub{
		updateTaskFunc: func(
			_ context.Context,
			_ int64,
			expectedVersion int64,
			patch core_domain.UpdateTask,
		) (core_domain.Task, error) {
			serviceExpectedVersion = expectedVersion
			servicePatch = patch
			return core_domain.NewTask(
				defaultTaskID,
				"Buy groceries",
				"",
				false,
				core_domain.DefaultTaskPriority,
				time.Now(),
				time.Now(),
				nil,
				responseVersion,
			), nil
		},
	}
	taskIDPath := strconv.FormatInt(defaultTaskID, 10)
	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/tasks/"+taskIDPath,
		strings.NewReader(`{"version":2,"due_at":null}`),
	)
	request.SetPathValue("id", taskIDPath)
	request = requestWithTestLogger(request)
	recorder := httptest.NewRecorder()

	NewTasksHTTPHandler(service).UpdateTask(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if !servicePatch.DueAt.Set {
		t.Fatal("expected due_at patch to be set")
	}
	if serviceExpectedVersion != requestVersion {
		t.Errorf(
			"expected service version %d, got %d",
			requestVersion,
			serviceExpectedVersion,
		)
	}
	if servicePatch.DueAt.Value != nil {
		t.Errorf("expected nil due_at value, got %v", servicePatch.DueAt.Value)
	}

	var response UpdateTaskResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.DueAt != nil {
		t.Errorf("expected cleared response due_at, got %v", response.DueAt)
	}
	if response.Version != responseVersion {
		t.Errorf("expected response version %d, got %d", responseVersion, response.Version)
	}
}

func TestTasksHTTPHandlerUpdateTaskInvalidRequest(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "missing version", body: `{"done":true}`},
		{name: "null version", body: `{"version":null,"done":true}`},
		{name: "zero version", body: `{"version":0,"done":true}`},
		{name: "negative version", body: `{"version":-1,"done":true}`},
		{name: "wrong version type", body: `{"version":"2","done":true}`},
		{name: "empty patch", body: `{"version":2}`},
		{name: "null title", body: `{"version":2,"title":null}`},
		{name: "null description", body: `{"version":2,"description":null}`},
		{name: "null done", body: `{"version":2,"done":null}`},
		{name: "null priority", body: `{"version":2,"priority":null}`},
		{name: "short title", body: `{"version":2,"title":"ab"}`},
		{name: "wrong done type", body: `{"version":2,"done":"true"}`},
		{name: "wrong priority type", body: `{"version":2,"priority":1}`},
		{name: "unsupported priority", body: `{"version":2,"priority":"critical"}`},
		{name: "invalid due_at", body: `{"version":2,"due_at":"tomorrow"}`},
		{name: "unknown field", body: `{"version":2,"completed":true}`},
		{name: "malformed JSON", body: `{"version":2,"done":`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceCalled := false
			service := tasksServiceStub{
				updateTaskFunc: func(
					_ context.Context,
					_ int64,
					_ int64,
					_ core_domain.UpdateTask,
				) (core_domain.Task, error) {
					serviceCalled = true
					return core_domain.Task{}, nil
				},
			}
			taskIDPath := strconv.FormatInt(defaultTaskID, 10)
			request := httptest.NewRequest(
				http.MethodPatch,
				"/api/v1/tasks/"+taskIDPath,
				strings.NewReader(tt.body),
			)
			request.SetPathValue("id", taskIDPath)
			request = requestWithTestLogger(request)
			recorder := httptest.NewRecorder()

			NewTasksHTTPHandler(service).UpdateTask(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
			}
			if serviceCalled {
				t.Fatal("service must not be called for an invalid request")
			}
			response := decodeErrorResponse(t, recorder)
			if response.Error != core_errors.ErrInvalidArgument.Error() {
				t.Errorf("expected error %q, got %q", core_errors.ErrInvalidArgument, response.Error)
			}
		})
	}
}

func TestTasksHTTPHandlerUpdateTaskInvalidID(t *testing.T) {
	serviceCalled := false
	service := tasksServiceStub{
		updateTaskFunc: func(
			_ context.Context,
			_ int64,
			_ int64,
			_ core_domain.UpdateTask,
		) (core_domain.Task, error) {
			serviceCalled = true
			return core_domain.Task{}, nil
		},
	}
	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/tasks/task",
		strings.NewReader(`{"version":2,"done":true}`),
	)
	request.SetPathValue("id", "task")
	request = requestWithTestLogger(request)
	recorder := httptest.NewRecorder()

	NewTasksHTTPHandler(service).UpdateTask(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if serviceCalled {
		t.Fatal("service must not be called for an invalid task ID")
	}
}

func TestTasksHTTPHandlerUpdateTaskNotFound(t *testing.T) {
	service := tasksServiceStub{
		updateTaskFunc: func(
			_ context.Context,
			_ int64,
			_ int64,
			_ core_domain.UpdateTask,
		) (core_domain.Task, error) {
			return core_domain.Task{}, core_errors.ErrNotFound
		},
	}
	taskIDPath := strconv.FormatInt(defaultTaskID, 10)
	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/tasks/"+taskIDPath,
		strings.NewReader(`{"version":2,"done":true}`),
	)
	request.SetPathValue("id", taskIDPath)
	request = requestWithTestLogger(request)
	recorder := httptest.NewRecorder()

	NewTasksHTTPHandler(service).UpdateTask(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
	response := decodeErrorResponse(t, recorder)
	if response.Error != core_errors.ErrNotFound.Error() {
		t.Errorf("expected error %q, got %q", core_errors.ErrNotFound, response.Error)
	}
}

func TestTasksHTTPHandlerUpdateTaskConflict(t *testing.T) {
	service := tasksServiceStub{
		updateTaskFunc: func(
			_ context.Context,
			_ int64,
			_ int64,
			_ core_domain.UpdateTask,
		) (core_domain.Task, error) {
			return core_domain.Task{}, core_errors.ErrConflict
		},
	}

	taskIDPath := strconv.FormatInt(defaultTaskID, 10)
	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/tasks/"+taskIDPath,
		strings.NewReader(`{"version":2,"done":true}`),
	)
	request.SetPathValue("id", taskIDPath)
	request = requestWithTestLogger(request)
	recorder := httptest.NewRecorder()

	NewTasksHTTPHandler(service).UpdateTask(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusConflict,
			recorder.Code,
			recorder.Body.String(),
		)
	}
	response := decodeErrorResponse(t, recorder)
	if response.Error != core_errors.ErrConflict.Error() {
		t.Errorf("expected error %q, got %q", core_errors.ErrConflict, response.Error)
	}
}

func TestTasksHTTPHandlerUpdateTaskServiceError(t *testing.T) {
	service := tasksServiceStub{
		updateTaskFunc: func(
			_ context.Context,
			_ int64,
			_ int64,
			_ core_domain.UpdateTask,
		) (core_domain.Task, error) {
			return core_domain.Task{}, errors.New(internalServiceErrorMessage)
		},
	}
	taskIDPath := strconv.FormatInt(defaultTaskID, 10)
	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/tasks/"+taskIDPath,
		strings.NewReader(`{"version":2,"done":true}`),
	)
	request.SetPathValue("id", taskIDPath)
	request = requestWithTestLogger(request)
	recorder := httptest.NewRecorder()

	NewTasksHTTPHandler(service).UpdateTask(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), internalServiceErrorMessage) {
		t.Error("internal service error must not be exposed in the response")
	}
	response := decodeErrorResponse(t, recorder)
	if response.Error != http.StatusText(http.StatusInternalServerError) {
		t.Errorf("expected error %q, got %q", http.StatusText(http.StatusInternalServerError), response.Error)
	}
}
