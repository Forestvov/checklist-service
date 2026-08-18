package tasks_transport_http

import (
	"net/http"

	core_logger "github.com/Forestvov/checklist-service/internal/core/logger"
	core_http_request "github.com/Forestvov/checklist-service/internal/core/transport/http/request"
	core_http_response "github.com/Forestvov/checklist-service/internal/core/transport/http/response"
)

type GetTasksResponse struct {
	Data []TaskDTOResponse                 `json:"data"`
	Meta core_http_response.PaginationMeta `json:"meta"`
}

// GetTasks returns a paginated list of tasks.
//
// @Summary Get tasks
// @Description Returns tasks ordered from newest to oldest using page-based pagination.
// @Description If page or per_page is omitted, the defaults are page=1 and per_page=20.
// @Description A page beyond the available range returns an empty data array with the requested pagination metadata.
// @Tags tasks
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param per_page query int false "Number of tasks per page" default(20) minimum(1) maximum(100)
// @Success 200 {object} GetTasksResponse "Paginated task list"
// @Failure 400 {object} core_http_response.ErrorResponse "Invalid pagination parameters"
// @Failure 500 {object} core_http_response.ErrorResponse "Unexpected server error"
// @Router /api/v1/tasks [get]
func (h *TasksHTTPHandler) GetTasks(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(logger, rw)

	paginationParams, err := core_http_request.GetPaginationParams(r)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get pagination parameters",
		)
		return
	}

	tasksResult, err := h.tasksService.GetTasks(
		ctx,
		paginationParams,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get tasks",
		)
		return
	}

	response := GetTasksResponse{
		Data: tasksDTOsFromDomain(tasksResult.Items),
		Meta: core_http_response.PaginationMetaFromResult(tasksResult),
	}
	responseHandler.JSONResponse(response, http.StatusOK)
}
