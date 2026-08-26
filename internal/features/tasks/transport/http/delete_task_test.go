package tasks_transport_http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

func TestTasksHTTPHandlerDeleteTaskSuccess(t *testing.T) {
	var serviceTaskID int64
	service := tasksServiceStub{
		deleteTaskFunc: func(
			_ context.Context,
			taskID int64,
		) error {
			serviceTaskID = taskID
			return nil
		},
	}

	handler := NewTasksHTTPHandler(service)

	taskIDPath := strconv.FormatInt(defaultTaskID, 10)

	request := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/tasks/"+taskIDPath,
		nil,
	)

	request.SetPathValue("id", taskIDPath)
	request = requestWithTestLogger(request)

	recorder := httptest.NewRecorder()

	handler.DeleteTask(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNoContent,
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

	if recorder.Body.Len() != 0 {
		t.Errorf(
			"expected empty response body, got %q",
			recorder.Body.String(),
		)
	}
}

func TestTasksHTTPHandlerDeleteTaskNotFound(t *testing.T) {
	service := tasksServiceStub{
		deleteTaskFunc: func(
			_ context.Context,
			_ int64,
		) error {
			return core_errors.ErrNotFound
		},
	}

	handler := NewTasksHTTPHandler(service)

	taskIDPath := strconv.FormatInt(defaultTaskID, 10)

	request := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/tasks/"+taskIDPath,
		nil,
	)

	request.SetPathValue("id", taskIDPath)
	request = requestWithTestLogger(request)
	recorder := httptest.NewRecorder()

	handler.DeleteTask(recorder, request)

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

func TestTasksHTTPHandlerDeleteTaskInvalidID(t *testing.T) {
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
				deleteTaskFunc: func(
					_ context.Context,
					_ int64,
				) error {
					serviceCalled = true
					return nil
				},
			}

			handler := NewTasksHTTPHandler(service)
			request := httptest.NewRequest(
				http.MethodDelete,
				"/api/v1/tasks/"+tt.value,
				nil,
			)
			if tt.value != "" {
				request.SetPathValue("id", tt.value)
			}

			request = requestWithTestLogger(request)
			recorder := httptest.NewRecorder()

			handler.DeleteTask(recorder, request)

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

func TestTasksHTTPHandlerDeleteTaskServiceError(t *testing.T) {
	service := tasksServiceStub{
		deleteTaskFunc: func(
			_ context.Context,
			_ int64,
		) error {
			return errors.New(internalServiceErrorMessage)
		},
	}

	handler := NewTasksHTTPHandler(service)

	taskIDPath := strconv.FormatInt(defaultTaskID, 10)

	request := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/tasks/"+taskIDPath,
		nil,
	)
	request.SetPathValue("id", taskIDPath)
	request = requestWithTestLogger(request)
	recorder := httptest.NewRecorder()

	handler.DeleteTask(recorder, request)

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
