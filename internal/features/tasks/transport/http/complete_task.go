package tasks_transport_http

import (
	"net/http"

	core_logger "github.com/Forestvov/checklist-service/internal/core/logger"
	core_http_request "github.com/Forestvov/checklist-service/internal/core/transport/http/request"
	core_http_response "github.com/Forestvov/checklist-service/internal/core/transport/http/response"
)

type CompleteTaskResponse TaskDTOResponse

// CompleteTask marks a task as completed.
func (h *TasksHTTPHandler) CompleteTask(
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

	taskDomain, err := h.tasksService.CompleteTask(ctx, taskID)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to complete task")
		return
	}

	response := CompleteTaskResponse(taskDTOFromDomain(taskDomain))
	responseHandler.JSONResponse(response, http.StatusOK)
}
