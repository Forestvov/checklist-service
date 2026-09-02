package tasks_transport_http

import (
	"fmt"
	"net/http"
	"time"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
	core_logger "github.com/Forestvov/checklist-service/internal/core/logger"
	core_http_request "github.com/Forestvov/checklist-service/internal/core/transport/http/request"
	core_http_response "github.com/Forestvov/checklist-service/internal/core/transport/http/response"
	core_http_types "github.com/Forestvov/checklist-service/internal/core/transport/http/types"
)

type UpdateTaskRequest struct {
	Title       core_http_types.Nullable[string]                   `json:"title" swaggertype:"string" minLength:"3" maxLength:"255" example:"Buy groceries"`
	Description core_http_types.Nullable[string]                   `json:"description" swaggertype:"string" maxLength:"5000" example:"Milk, bread and eggs"`
	Done        core_http_types.Nullable[bool]                     `json:"done" swaggertype:"boolean" example:"true"`
	Priority    core_http_types.Nullable[core_domain.TaskPriority] `json:"priority" swaggertype:"string" enums:"low,medium,high" example:"medium"`
	DueAt       core_http_types.Nullable[time.Time]                `json:"due_at" swaggertype:"string" format:"date-time" example:"2026-09-15T12:00:00Z"`
	Version     *int64                                             `json:"version" validate:"required" minimum:"1" example:"3"`
}

func (r *UpdateTaskRequest) Validate() error {
	if r.Version == nil {
		return fmt.Errorf(
			"version is required: %w",
			core_errors.ErrInvalidArgument,
		)
	}

	if *r.Version <= 0 {
		return fmt.Errorf(
			"version must be greater than zero: %w",
			core_errors.ErrInvalidArgument,
		)
	}

	if err := taskUpdateFromRequest(*r).Validate(); err != nil {
		return fmt.Errorf("validate task update: %w", err)
	}

	return nil
}

type UpdateTaskResponse TaskDTOResponse

// UpdateTask partially updates a task.
//
// @Summary Update a task
// @Description Partially updates the task with the specified integer identifier.
// @Description Version is required and must match the current task version.
// @Description A successful update increments version by one.
// @Description Send at least one of title, description, done, priority, or due_at.
// @Description Omitted fields remain unchanged.
// @Description Null is allowed only for due_at and removes the current deadline.
// @Description The title must contain 3–255 Unicode characters after trimming. The description must not exceed 5000 Unicode characters.
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path int true "Task identifier"
// @Param payload body UpdateTaskRequest true "Fields to update and expected version"
// @Success 200 {object} UpdateTaskResponse "Task updated successfully"
// @Failure 400 {object} core_http_response.ErrorResponse "Malformed JSON, invalid task identifier or version, or request validation failed"
// @Failure 404 {object} core_http_response.ErrorResponse "Task not found"
// @Failure 409 {object} core_http_response.ErrorResponse "Task version conflict"
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

	taskDomain, err := h.tasksService.UpdateTask(
		ctx,
		taskID,
		*req.Version,
		taskUpdate,
	)
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
		req.Priority.ToDomain(),
		req.DueAt.ToDomain(),
	)
}
