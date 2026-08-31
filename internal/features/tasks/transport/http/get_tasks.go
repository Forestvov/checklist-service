package tasks_transport_http

import (
	"fmt"
	"net/http"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
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
// @Description Returns a filtered and sorted task list using page-based pagination.
// @Description If page or per_page is omitted, the defaults are page=1 and per_page=20.
// @Description If done is omitted, tasks of both statuses are returned.
// @Description If priority is omitted, tasks of all priorities are returned.
// @Description By default, tasks are ordered by created_at in descending order.
// @Description A page beyond the available range returns an empty data array with the requested pagination metadata.
// @Tags tasks
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param per_page query int false "Number of tasks per page" default(20) minimum(1) maximum(100)
// @Param done query bool false "Filter by completion status"
// @Param priority query string false "Filter by priority" Enums(low,medium,high)
// @Param sort query string false "Sort field" Enums(created_at,updated_at,title,priority) default(created_at)
// @Param order query string false "Sort direction" Enums(asc,desc) default(desc)
// @Success 200 {object} GetTasksResponse "Paginated task list"
// @Failure 400 {object} core_http_response.ErrorResponse "Invalid pagination or filter parameters"
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

	done, priority, sortBy, orderBy, err := getTasksFilterQueryParams(r)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get task filter parameters",
		)
		return
	}

	filter := core_domain.NewTaskFilter(
		done,
		priority,
		sortBy,
		orderBy,
	)

	tasksResult, err := h.tasksService.GetTasks(
		ctx,
		paginationParams,
		filter,
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

func getTasksFilterQueryParams(r *http.Request) (*bool, *core_domain.TaskPriority, core_domain.TaskSort, core_domain.SortOrder, error) {
	const (
		doneParam     = "done"
		priorityParam = "priority"
		sortByParam   = "sort"
		orderByParam  = "order"
	)

	done, err := core_http_request.GetBoolQueryParam(r, doneParam)
	if err != nil {
		return nil, nil, "", "", fmt.Errorf("failed to get done param: %w", err)
	}

	query := r.URL.Query()
	priority, err := getTaskPriorityQueryParam(query, priorityParam)
	if err != nil {
		return nil, nil, "", "", fmt.Errorf("failed to get priority param: %w", err)
	}

	sortBy := core_domain.TaskSort(query.Get(sortByParam))
	order := core_domain.SortOrder(query.Get(orderByParam))

	return done, priority, sortBy, order, nil
}

func getTaskPriorityQueryParam(
	query map[string][]string,
	key string,
) (*core_domain.TaskPriority, error) {
	values, exists := query[key]
	if !exists {
		return nil, nil
	}
	if len(values) != 1 || values[0] == "" {
		return nil, fmt.Errorf(
			"query parameter %q must contain exactly one non-empty value: %w",
			key,
			core_errors.ErrInvalidArgument,
		)
	}

	priority := core_domain.TaskPriority(values[0])
	return &priority, nil
}
