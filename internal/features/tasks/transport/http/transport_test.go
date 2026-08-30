package tasks_transport_http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_logger "github.com/Forestvov/checklist-service/internal/core/logger"
	core_pagination "github.com/Forestvov/checklist-service/internal/core/pagination"
	core_http_response "github.com/Forestvov/checklist-service/internal/core/transport/http/response"
	"go.uber.org/zap"
)

const (
	defaultTaskID               int64 = 42
	internalServiceErrorMessage       = "database password=secret"
)

type tasksServiceStub struct {
	createTaskFunc func(
		ctx context.Context,
		task core_domain.Task,
	) (core_domain.Task, error)

	getTasksFunc func(
		ctx context.Context,
		paginationParams core_pagination.Params,
		filter core_domain.TaskFilter,
	) (core_pagination.Result[core_domain.Task], error)

	getTaskFunc func(
		ctx context.Context,
		taskID int64,
	) (core_domain.Task, error)

	updateTaskFunc func(
		ctx context.Context,
		taskID int64,
		taskUpdate core_domain.UpdateTask,
	) (core_domain.Task, error)

	deleteTaskFunc func(
		ctx context.Context,
		taskID int64,
	) error
}

func TestTasksHTTPHandlerRoutes(t *testing.T) {
	handler := NewTasksHTTPHandler(tasksServiceStub{})

	routes := handler.Routes()

	expectedRoutes := map[string]struct{}{
		http.MethodPost + " /tasks":        {},
		http.MethodGet + " /tasks":         {},
		http.MethodGet + " /tasks/{id}":    {},
		http.MethodPatch + " /tasks/{id}":  {},
		http.MethodDelete + " /tasks/{id}": {},
	}

	if len(routes) != len(expectedRoutes) {
		t.Fatalf(
			"expected %d routes, got %d",
			len(expectedRoutes),
			len(routes),
		)
	}

	for _, route := range routes {
		key := route.Method + " " + route.Path

		if _, ok := expectedRoutes[key]; !ok {
			t.Errorf("unexpected route %q", key)
			continue
		}

		if route.Handler == nil {
			t.Errorf("route %q has nil handler", key)
		}

		delete(expectedRoutes, key)
	}

	for missingRoute := range expectedRoutes {
		t.Errorf("missing route %q", missingRoute)
	}
}

func (s tasksServiceStub) CreateTask(
	ctx context.Context,
	task core_domain.Task,
) (core_domain.Task, error) {
	return s.createTaskFunc(ctx, task)
}

func (s tasksServiceStub) GetTasks(
	ctx context.Context,
	paginationParams core_pagination.Params,
	filter core_domain.TaskFilter,
) (core_pagination.Result[core_domain.Task], error) {
	return s.getTasksFunc(
		ctx,
		paginationParams,
		filter,
	)
}

func (s tasksServiceStub) GetTask(
	ctx context.Context,
	taskID int64,
) (core_domain.Task, error) {
	return s.getTaskFunc(ctx, taskID)
}

func (s tasksServiceStub) UpdateTask(
	ctx context.Context,
	taskID int64,
	taskUpdate core_domain.UpdateTask,
) (core_domain.Task, error) {
	return s.updateTaskFunc(ctx, taskID, taskUpdate)
}

func (s tasksServiceStub) DeleteTask(
	ctx context.Context,
	taskID int64,
) error {
	return s.deleteTaskFunc(ctx, taskID)
}

func requestWithTestLogger(r *http.Request) *http.Request {
	logger := &core_logger.Logger{
		Logger: zap.NewNop(),
	}

	ctx := core_logger.ToContext(r.Context(), logger)

	return r.WithContext(ctx)
}

func decodeErrorResponse(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
) core_http_response.ErrorResponse {
	t.Helper()

	var response core_http_response.ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}

	return response
}
