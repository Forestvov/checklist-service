package tasks_transport_http

import (
	"net/http"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_logger "github.com/Forestvov/checklist-service/internal/core/logger"
	core_http_request "github.com/Forestvov/checklist-service/internal/core/transport/http/request"
	core_http_response "github.com/Forestvov/checklist-service/internal/core/transport/http/response"
	core_http_types "github.com/Forestvov/checklist-service/internal/core/transport/http/types"
)

type UpdateTaskRequest struct {
	Title       core_http_types.Nullable[string] `json:"title" swaggertype:"string" minLength:"3" maxLength:"255" example:"Buy groceries"`
	Description core_http_types.Nullable[string] `json:"description" swaggertype:"string" maxLength:"5000" example:"Milk, bread and eggs"`
	Done        core_http_types.Nullable[bool]   `json:"done" swaggertype:"boolean" example:"true"`
}

func (r *UpdateTaskRequest) Validate() error {
	return taskUpdateFromRequest(*r).Validate()
}

type UpdateTaskResponse TaskDTOResponse

// UpdateTask partially updates a task.
//
// @Summary Update a task
// @Description Partially updates the task with the specified integer identifier.
// @Description Send at least one of title, description, or done. Omitted fields remain unchanged; null values are not allowed.
// @Description The title must contain 3–255 Unicode characters after leading and trailing whitespace is ignored. The description must not exceed 5000 Unicode characters.
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path int true "Task identifier"
// @Param payload body UpdateTaskRequest true "Fields to update"
// @Success 200 {object} UpdateTaskResponse "Task updated successfully"
// @Failure 400 {object} core_http_response.ErrorResponse "Malformed JSON, invalid task identifier, or request validation failed"
// @Failure 404 {object} core_http_response.ErrorResponse "Task not found"
// @Failure 500 {object} core_http_response.ErrorResponse "Unexpected server error"
// @Router /api/v1/tasks/{id} [patch]
func (h *TasksHTTPHandler) UpdateTask(
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

	var req UpdateTaskRequest
	if err := core_http_request.DecodeValidateRequest(r, &req); err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to decode and validate HTTP request",
		)
		return
	}

	taskUpdate := taskUpdateFromRequest(req)

	taskDomain, err := h.tasksService.UpdateTask(ctx, taskID, taskUpdate)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to patch task",
		)
		return
	}

	response := UpdateTaskResponse(taskDTOFromDomain(taskDomain))

	responseHandler.JSONResponse(response, http.StatusOK)
}

func taskUpdateFromRequest(req UpdateTaskRequest) core_domain.UpdateTask {
	return core_domain.NewUpdateTask(
		req.Title.ToDomain(),
		req.Description.ToDomain(),
		req.Done.ToDomain(),
	)
}
