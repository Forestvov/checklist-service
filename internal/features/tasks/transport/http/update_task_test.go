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
	createdAt := time.Date(2026, time.December, 25, 0, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	expectedTask := core_domain.NewTask(
		defaultTaskID,
		"Buy groceries",
		"Milk, bread and eggs",
		true,
		createdAt,
		updatedAt,
	)

	var (
		serviceTaskID int64
		servicePatch  core_domain.UpdateTask
	)
	service := tasksServiceStub{
		updateTaskFunc: func(
			_ context.Context,
			taskID int64,
			patch core_domain.UpdateTask,
		) (core_domain.Task, error) {
			serviceTaskID = taskID
			servicePatch = patch
			return expectedTask, nil
		},
	}

	taskIDPath := strconv.FormatInt(defaultTaskID, 10)
	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/tasks/"+taskIDPath,
		strings.NewReader(`{"description":"Milk, bread and eggs","done":true}`),
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
	if servicePatch.Title.Set {
		t.Error("expected omitted title to remain unset")
	}
	if !servicePatch.Description.Set || servicePatch.Description.Value == nil || *servicePatch.Description.Value != expectedTask.Description {
		t.Errorf("unexpected description patch: %+v", servicePatch.Description)
	}
	if !servicePatch.Done.Set || servicePatch.Done.Value == nil || !*servicePatch.Done.Value {
		t.Errorf("unexpected done patch: %+v", servicePatch.Done)
	}

	var response UpdateTaskResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ID != expectedTask.ID ||
		response.Title != expectedTask.Title ||
		response.Description != expectedTask.Description ||
		response.Done != expectedTask.Done ||
		!response.CreatedAt.Equal(expectedTask.CreatedAt) ||
		!response.UpdatedAt.Equal(expectedTask.UpdatedAt) {
		t.Errorf("expected response for task %+v, got %+v", expectedTask, response)
	}
}

func TestTasksHTTPHandlerUpdateTaskInvalidRequest(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "empty patch", body: `{}`},
		{name: "null title", body: `{"title":null}`},
		{name: "null description", body: `{"description":null}`},
		{name: "null done", body: `{"done":null}`},
		{name: "short title", body: `{"title":"ab"}`},
		{name: "wrong done type", body: `{"done":"true"}`},
		{name: "unknown field", body: `{"completed":true}`},
		{name: "malformed JSON", body: `{"done":`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceCalled := false
			service := tasksServiceStub{
				updateTaskFunc: func(
					_ context.Context,
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
			_ core_domain.UpdateTask,
		) (core_domain.Task, error) {
			serviceCalled = true
			return core_domain.Task{}, nil
		},
	}
	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/tasks/task",
		strings.NewReader(`{"done":true}`),
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
			_ core_domain.UpdateTask,
		) (core_domain.Task, error) {
			return core_domain.Task{}, core_errors.ErrNotFound
		},
	}
	taskIDPath := strconv.FormatInt(defaultTaskID, 10)
	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/tasks/"+taskIDPath,
		strings.NewReader(`{"done":true}`),
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

func TestTasksHTTPHandlerUpdateTaskServiceError(t *testing.T) {
	service := tasksServiceStub{
		updateTaskFunc: func(
			_ context.Context,
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
		strings.NewReader(`{"done":true}`),
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
