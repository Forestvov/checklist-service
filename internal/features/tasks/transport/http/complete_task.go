package tasks_transport_http

import (
	"net/http"

	core_logger "github.com/Forestvov/checklist-service/internal/core/logger"
	core_http_request "github.com/Forestvov/checklist-service/internal/core/transport/http/request"
	core_http_response "github.com/Forestvov/checklist-service/internal/core/transport/http/response"
)

type CompleteTaskResponse TaskDTOResponse

// CompleteTask marks a task as completed.
//
// @Summary Complete a task
// @Description Marks the task with the specified integer identifier as completed.
// @Description Returns the updated task with done=true and a refreshed updated_at timestamp.
// @Tags tasks
// @Produce json
// @Param id path int true "Task identifier"
// @Success 200 {object} CompleteTaskResponse "Task completed successfully"
// @Failure 400 {object} core_http_response.ErrorResponse "Task identifier is not a valid integer"
// @Failure 404 {object} core_http_response.ErrorResponse "Task not found"
// @Failure 500 {object} core_http_response.ErrorResponse "Unexpected server error"
// @Router /api/v1/tasks/{id} [patch]
func (h *TasksHTTPHandler) CompleteTask(
	rw http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(logger, rw)

	taskId, err := core_http_request.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get taskId path value")
		return
	}

	taskDomain, err := h.tasksService.CompleteTask(ctx, int64(taskId))
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to complete task")
		return
	}

	response := CompleteTaskResponse(taskDTOFromDomain(taskDomain))
	responseHandler.JSONResponse(response, http.StatusOK)
}
