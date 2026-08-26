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

func TestTasksHTTPHandlerCompleteTaskSuccess(t *testing.T) {
	createdAt := time.Date(
		2026,
		time.December,
		25,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	updatedAt := createdAt.Add(time.Hour)

	expectedTask := core_domain.NewTask(
		defaultTaskID,
		"Buy groceries",
		"Milk and bread",
		true,
		createdAt,
		updatedAt,
	)

	var serviceTaskID int64
	service := tasksServiceStub{
		completeTaskFunc: func(
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
		http.MethodPatch,
		"/api/v1/tasks/"+taskIDPath,
		nil,
	)

	request.SetPathValue("id", taskIDPath)
	request = requestWithTestLogger(request)

	recorder := httptest.NewRecorder()

	handler.CompleteTask(recorder, request)

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

	var response CompleteTaskResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.ID != expectedTask.ID {
		t.Errorf(
			"expected task ID %d, got %d",
			expectedTask.ID,
			response.ID,
		)
	}

	if !response.Done {
		t.Error("expected completed task")
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

	if !response.CreatedAt.Equal(createdAt) {
		t.Errorf(
			"expected created_at %v, got %v",
			createdAt,
			response.CreatedAt,
		)
	}

	if !response.UpdatedAt.Equal(updatedAt) {
		t.Errorf(
			"expected updated_at %v, got %v",
			updatedAt,
			response.UpdatedAt,
		)
	}
}

func TestTasksHTTPHandlerCompleteTaskNotFound(t *testing.T) {
	service := tasksServiceStub{
		completeTaskFunc: func(
			_ context.Context,
			_ int64,
		) (core_domain.Task, error) {
			return core_domain.Task{}, core_errors.ErrNotFound
		},
	}

	handler := NewTasksHTTPHandler(service)

	taskIDPath := strconv.FormatInt(defaultTaskID, 10)

	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/tasks/"+taskIDPath,
		nil,
	)

	request.SetPathValue("id", taskIDPath)
	request = requestWithTestLogger(request)
	recorder := httptest.NewRecorder()

	handler.CompleteTask(recorder, request)

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

func TestTasksHTTPHandlerCompleteTaskInvalidID(t *testing.T) {
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
				completeTaskFunc: func(
					_ context.Context,
					_ int64,
				) (core_domain.Task, error) {
					serviceCalled = true
					return core_domain.Task{}, nil
				},
			}

			handler := NewTasksHTTPHandler(service)
			request := httptest.NewRequest(
				http.MethodPatch,
				"/api/v1/tasks/"+tt.value,
				nil,
			)
			if tt.value != "" {
				request.SetPathValue("id", tt.value)
			}

			request = requestWithTestLogger(request)
			recorder := httptest.NewRecorder()

			handler.CompleteTask(recorder, request)

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

func TestTasksHTTPHandlerCompleteTaskServiceError(t *testing.T) {
	service := tasksServiceStub{
		completeTaskFunc: func(
			_ context.Context,
			_ int64,
		) (core_domain.Task, error) {
			return core_domain.Task{}, errors.New(internalServiceErrorMessage)
		},
	}

	handler := NewTasksHTTPHandler(service)

	taskIDPath := strconv.FormatInt(defaultTaskID, 10)

	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/tasks/"+taskIDPath,
		nil,
	)
	request.SetPathValue("id", taskIDPath)
	request = requestWithTestLogger(request)
	recorder := httptest.NewRecorder()

	handler.CompleteTask(recorder, request)

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
