package tasks_transport_http

import (
	"net/http"

	core_logger "github.com/Forestvov/checklist-service/internal/core/logger"
	core_http_request "github.com/Forestvov/checklist-service/internal/core/transport/http/request"
	core_http_response "github.com/Forestvov/checklist-service/internal/core/transport/http/response"
)

// DeleteTask deletes a task by its identifier.
//
// @Summary Delete a task
// @Description Permanently deletes the task with the specified integer identifier.
// @Description A successful deletion returns 204 with an empty response body.
// @Tags tasks
// @Param id path int true "Task identifier"
// @Success 204 "Task deleted successfully"
// @Failure 400 {object} core_http_response.ErrorResponse "Task identifier is not a valid integer"
// @Failure 404 {object} core_http_response.ErrorResponse "Task not found"
// @Failure 500 {object} core_http_response.ErrorResponse "Unexpected server error"
// @Router /api/v1/tasks/{id} [delete]
func (h *TasksHTTPHandler) DeleteTask(
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

	if err := h.tasksService.DeleteTask(
		ctx,
		taskID,
	); err != nil {
		responseHandler.ErrorResponse(err, "failed to delete task")
		return
	}

	responseHandler.NoContentResponse()
}
