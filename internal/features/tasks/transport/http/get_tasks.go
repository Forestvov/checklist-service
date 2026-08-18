package tasks_transport_http

import (
	"net/http"

	core_logger "github.com/Forestvov/checklist-service/internal/core/logger"
	core_http_response "github.com/Forestvov/checklist-service/internal/core/transport/http/response"
)

func (h *TasksHTTPHandler) GetTasks(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(logger, rw)

	tasksDomains, err := h.tasksService.GetTasks(ctx)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get tasks",
		)
		return
	}

	response := tasksDTOsFromDomain(tasksDomains)
	responseHandler.JSONResponse(response, http.StatusOK)
}
