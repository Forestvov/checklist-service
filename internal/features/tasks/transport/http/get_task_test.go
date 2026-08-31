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

func TestTasksHTTPHandlerGetTaskSuccess(t *testing.T) {
	now := time.Date(2026, time.December, 25, 0, 0, 0, 0, time.UTC)

	expectedTask := core_domain.NewTask(
		defaultTaskID,
		"Buy groceries",
		"Milk and bread",
		false,
		core_domain.DefaultTaskPriority,
		now,
		now,
	)

	var serviceTaskID int64
	service := tasksServiceStub{
		getTaskFunc: func(
			_ context.Context,
			taskID int64,
		) (core_domain.Task, error) {
			serviceTaskID = taskID
			return expectedTask, nil
		},
	}

	handler := NewTasksHTTPHandler(service)

	taskIDPath := strconv.FormatInt(defaultTaskID, 10)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/tasks/"+taskIDPath,
		nil,
	)

	request.SetPathValue("id", taskIDPath)
	request = requestWithTestLogger(request)

	recorder := httptest.NewRecorder()

	handler.GetTask(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	if serviceTaskID != defaultTaskID {
		t.Errorf(
			"expected service task ID %d, got %d",
			defaultTaskID,
			serviceTaskID,
		)
	}

	var response GetTaskResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decoding response body: %v", err)
	}

	if response.ID != expectedTask.ID {
		t.Errorf(
			"expected response task ID %d, got %d",
			expectedTask.ID,
			response.ID,
		)
	}

	if response.Title != expectedTask.Title {
		t.Errorf(
			"expected title %q, got %q",
			expectedTask.Title,
			response.Title,
		)
	}

	if response.Description != expectedTask.Description {
		t.Errorf(
			"expected description %q, got %q",
			expectedTask.Description,
			response.Description,
		)
	}

	if response.Done != expectedTask.Done {
		t.Errorf(
			"expected done=%t, got %t",
			expectedTask.Done,
			response.Done,
		)
	}

	if response.Priority != expectedTask.Priority {
		t.Errorf(
			"expected priority %q, got %q",
			expectedTask.Priority,
			response.Priority,
		)
	}
}

func TestTasksHTTPHandlerGetTaskNotFound(t *testing.T) {

	service := tasksServiceStub{
		getTaskFunc: func(
			_ context.Context,
			_ int64,
		) (core_domain.Task, error) {
			return core_domain.Task{}, core_errors.ErrNotFound
		},
	}

	handler := NewTasksHTTPHandler(service)

	taskIDPath := strconv.FormatInt(defaultTaskID, 10)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/tasks/"+taskIDPath,
		nil,
	)

	request.SetPathValue("id", taskIDPath)
	request = requestWithTestLogger(request)
	recorder := httptest.NewRecorder()

	handler.GetTask(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			recorder.Code,
		)
	}

	response := decodeErrorResponse(t, recorder)
	if response.Error != core_errors.ErrNotFound.Error() {
		t.Errorf(
			"expected error %q, got %q",
			core_errors.ErrNotFound.Error(),
			response.Error,
		)
	}
}

func TestTasksHTTPHandlerGetTaskInvalidID(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{name: "missing ID", value: ""},
		{name: "zero ID", value: "0"},
		{name: "negative ID", value: "-1"},
		{name: "non-integer ID", value: "task"},
		{name: "overflow", value: "9223372036854775808"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceCalled := false
			service := tasksServiceStub{
				getTaskFunc: func(
					_ context.Context,
					_ int64,
				) (core_domain.Task, error) {
					serviceCalled = true
					return core_domain.Task{}, nil
				},
			}

			handler := NewTasksHTTPHandler(service)
			request := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/tasks/"+tt.value,
				nil,
			)
			if tt.value != "" {
				request.SetPathValue("id", tt.value)
			}
			request = requestWithTestLogger(request)
			recorder := httptest.NewRecorder()

			handler.GetTask(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf(
					"expected status %d, got %d",
					http.StatusBadRequest,
					recorder.Code,
				)
			}
			if serviceCalled {
				t.Fatal("service must not be called for an invalid task ID")
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

func TestTasksHTTPHandlerGetTaskServiceError(t *testing.T) {
	service := tasksServiceStub{
		getTaskFunc: func(
			_ context.Context,
			_ int64,
		) (core_domain.Task, error) {
			return core_domain.Task{}, errors.New(internalServiceErrorMessage)
		},
	}

	handler := NewTasksHTTPHandler(service)

	taskIDPath := strconv.FormatInt(defaultTaskID, 10)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/tasks/"+taskIDPath,
		nil,
	)
	request.SetPathValue("id", taskIDPath)
	request = requestWithTestLogger(request)
	recorder := httptest.NewRecorder()

	handler.GetTask(recorder, request)

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
