package tasks_transport_http

import (
	"net/http"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_logger "github.com/Forestvov/checklist-service/internal/core/logger"
	core_http_request "github.com/Forestvov/checklist-service/internal/core/transport/http/request"
	core_http_response "github.com/Forestvov/checklist-service/internal/core/transport/http/response"
)

type CreateTaskRequest struct {
	// Title is required and must contain between 3 and 255 Unicode characters
	// after leading and trailing whitespace is ignored.
	Title string `json:"title" validate:"required,min=3,max=255" minLength:"3" maxLength:"255" example:"Buy groceries"`
	// Description is optional and may contain up to 5000 Unicode characters.
	Description *string `json:"description" validate:"omitempty,max=5000" maxLength:"5000" example:"Milk, bread, and vegetables"`
}

type CreateTaskResponse TaskDTOResponse

func (request CreateTaskRequest) Validate() error {
	task := core_domain.NewTaskUninitialized(
		request.Title,
		request.Description,
	)

	return task.Validate()
}

// CreateTask creates a new task.
//
// @Summary Create a task
// @Description Creates a new task with a required title and an optional description.
// @Description The title must contain 3–255 Unicode characters after leading and trailing whitespace is ignored.
// @Description The description may be omitted and must not exceed 5000 Unicode characters when provided.
// @Tags tasks
// @Accept json
// @Produce json
// @Param payload body CreateTaskRequest true "Task creation data"
// @Success 201 {object} CreateTaskResponse "Task created successfully"
// @Failure 400 {object} core_http_response.ErrorResponse "Malformed JSON or request validation failed"
// @Failure 500 {object} core_http_response.ErrorResponse "Unexpected server error"
// @Router /api/v1/tasks [post]
func (h *TasksHTTPHandler) CreateTask(
	rw http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	logger := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(logger, rw)

	var request CreateTaskRequest
	if err := core_http_request.DecodeValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to decode and validate HTTP request",
		)
		return
	}

	taskDomain := core_domain.NewTaskUninitialized(
		request.Title,
		request.Description,
	)

	taskDomain, err := h.tasksService.CreateTask(ctx, taskDomain)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to create task",
		)
		return
	}

	response := CreateTaskResponse(taskDTOFromDomain(taskDomain))
	responseHandler.JSONResponse(response, http.StatusCreated)
}
