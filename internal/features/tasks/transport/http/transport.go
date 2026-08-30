package tasks_transport_http

import (
	"context"
	"net/http"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_pagination "github.com/Forestvov/checklist-service/internal/core/pagination"
	core_http_server "github.com/Forestvov/checklist-service/internal/core/transport/http/server"
)

type TasksHTTPHandler struct {
	tasksService TasksService
}

type TasksService interface {
	CreateTask(
		ctx context.Context,
		task core_domain.Task,
	) (core_domain.Task, error)

	GetTasks(
		ctx context.Context,
		paginationParams core_pagination.Params,
		filter core_domain.TaskFilter,
	) (core_pagination.Result[core_domain.Task], error)

	GetTask(
		ctx context.Context,
		taskID int64,
	) (core_domain.Task, error)

	UpdateTask(
		ctx context.Context,
		taskID int64,
		taskUpdate core_domain.UpdateTask,
	) (core_domain.Task, error)

	DeleteTask(
		ctx context.Context,
		taskID int64,
	) error
}

func NewTasksHTTPHandler(tasksService TasksService) *TasksHTTPHandler {
	return &TasksHTTPHandler{
		tasksService: tasksService,
	}
}

func (h *TasksHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/tasks",
			Handler: h.CreateTask,
		},
		{
			Method:  http.MethodGet,
			Path:    "/tasks",
			Handler: h.GetTasks,
		},
		{
			Method:  http.MethodGet,
			Path:    "/tasks/{id}",
			Handler: h.GetTask,
		},
		{
			Method:  http.MethodPatch,
			Path:    "/tasks/{id}",
			Handler: h.UpdateTask,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/tasks/{id}",
			Handler: h.DeleteTask,
		},
	}
}
