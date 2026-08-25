package tasks_transport_http

import (
	"net/http"

	core_logger "github.com/Forestvov/checklist-service/internal/core/logger"
	core_http_request "github.com/Forestvov/checklist-service/internal/core/transport/http/request"
	core_http_response "github.com/Forestvov/checklist-service/internal/core/transport/http/response"
)

type GetTaskResponse TaskDTOResponse

// GetTask returns a task by its identifier.
//
// @Summary Get a task
// @Description Returns a single task with the specified integer identifier.
// @Description Returns 404 when the task does not exist.
// @Tags tasks
// @Produce json
// @Param id path int true "Task identifier"
// @Success 200 {object} GetTaskResponse "Task found"
// @Failure 400 {object} core_http_response.ErrorResponse "Task identifier is not a valid integer"
// @Failure 404 {object} core_http_response.ErrorResponse "Task not found"
// @Failure 500 {object} core_http_response.ErrorResponse "Unexpected server error"
// @Router /api/v1/tasks/{id} [get]
func (h *TasksHTTPHandler) GetTask(
	rw http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(logger, rw)

	taskID, err := core_http_request.GetInt64PathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get task ID path value")
		return
	}

	taskDomain, err := h.tasksService.GetTask(ctx, taskID)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get task")
		return
	}

	response := GetTaskResponse(taskDTOFromDomain(taskDomain))
	responseHandler.JSONResponse(response, http.StatusOK)
}
