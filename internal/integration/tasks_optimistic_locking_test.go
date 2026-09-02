//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
	core_logger "github.com/Forestvov/checklist-service/internal/core/logger"
	core_http_middleware "github.com/Forestvov/checklist-service/internal/core/transport/http/middleware"
	core_http_response "github.com/Forestvov/checklist-service/internal/core/transport/http/response"
	core_http_server "github.com/Forestvov/checklist-service/internal/core/transport/http/server"
	tasks_postgres_repository "github.com/Forestvov/checklist-service/internal/features/tasks/repository/postgres"
	tasks_service "github.com/Forestvov/checklist-service/internal/features/tasks/service"
	tasks_transport_http "github.com/Forestvov/checklist-service/internal/features/tasks/transport/http"
	postgres_testutil "github.com/Forestvov/checklist-service/internal/testutil/postgres"
	"go.uber.org/zap"
)

const (
	e2eSetupTimeout   = time.Minute * 2
	e2eRequestTimeout = time.Second * 5
)

func TestTasksOptimisticLockingE2E(t *testing.T) {
	setupContext, cancelSetup := context.WithTimeout(
		t.Context(),
		e2eSetupTimeout,
	)

	database, err := postgres_testutil.Start(setupContext)
	cancelSetup()

	if err != nil {
		t.Fatalf("start test PostgreSQL: %v", err)
	}

	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("closing PostgreSQL database: %v", err)
		}
	})

	server := newTasksTestServer(t, database)
	client := server.Client()

	createHTTPResponse := doJSONRequest(
		t,
		client,
		http.MethodPost,
		server.URL+"/api/v1/tasks",
		`{"title":"Optimistic locking task"}`,
	)

	assertHTTPStatus(t, createHTTPResponse, http.StatusCreated)

	createdTask := decodeJSON[tasks_transport_http.CreateTaskResponse](
		t,
		createHTTPResponse.Body,
	)

	if createdTask.ID <= 0 {
		t.Fatalf("expected positive task ID, got %d", createdTask.ID)
	}

	if createdTask.Version != 1 {
		t.Fatalf(
			"expected created task version 1, got %d",
			createdTask.Version,
		)
	}

	taskURL := fmt.Sprintf(
		"%s/api/v1/tasks/%d",
		server.URL,
		createdTask.ID,
	)

	getHTTPResponse := doJSONRequest(
		t,
		client,
		http.MethodGet,
		taskURL,
		"",
	)

	assertHTTPStatus(t, getHTTPResponse, http.StatusOK)

	taskBeforeUpdate := decodeJSON[tasks_transport_http.GetTaskResponse](
		t,
		getHTTPResponse.Body,
	)

	if taskBeforeUpdate.Version != 1 {
		t.Fatalf(
			"expected task version 1 before update, got %d",
			taskBeforeUpdate.Version,
		)
	}

	updatePayload := fmt.Sprintf(
		`{"version":%d,"title":"Updated title"}`,
		taskBeforeUpdate.Version,
	)

	updateHTTPResponse := doJSONRequest(
		t,
		client,
		http.MethodPatch,
		taskURL,
		updatePayload,
	)

	assertHTTPStatus(t, updateHTTPResponse, http.StatusOK)

	updatedTask := decodeJSON[tasks_transport_http.UpdateTaskResponse](
		t,
		updateHTTPResponse.Body,
	)

	if updatedTask.Title != "Updated title" {
		t.Errorf(
			"expected updated title %q, got %q",
			"Updated title",
			updatedTask.Title,
		)
	}

	if updatedTask.Version != 2 {
		t.Errorf(
			"expected task version 2 after update, got %d",
			updatedTask.Version,
		)
	}

	staleUpdatePayload := fmt.Sprintf(
		`{"version":%d,"done":true}`,
		createdTask.Version,
	)

	conflictHTTPResponse := doJSONRequest(
		t,
		client,
		http.MethodPatch,
		taskURL,
		staleUpdatePayload,
	)

	assertHTTPStatus(t, conflictHTTPResponse, http.StatusConflict)

	conflictResponse := decodeJSON[core_http_response.ErrorResponse](
		t,
		conflictHTTPResponse.Body,
	)

	if conflictResponse.Error != core_errors.ErrConflict.Error() {
		t.Errorf(
			"expected error %q, got %q",
			core_errors.ErrConflict.Error(),
			conflictResponse.Error,
		)
	}

	finalGetHTTPResponse := doJSONRequest(
		t,
		client,
		http.MethodGet,
		taskURL,
		"",
	)

	assertHTTPStatus(t, finalGetHTTPResponse, http.StatusOK)

	finalTask := decodeJSON[tasks_transport_http.GetTaskResponse](
		t,
		finalGetHTTPResponse.Body,
	)

	if finalTask.Version != 2 {
		t.Errorf(
			"expected final version 2, got %d",
			finalTask.Version,
		)
	}

	if finalTask.Title != "Updated title" {
		t.Errorf(
			"expected final title %q, got %q",
			"Updated title",
			finalTask.Title,
		)
	}

	if finalTask.Done {
		t.Error("stale update changed done to true")
	}
}

type testHTTPResponse struct {
	StatusCode int
	Body       []byte
}

func newTasksTestServer(
	t *testing.T,
	database *postgres_testutil.Database,
) *httptest.Server {
	t.Helper()

	tasksRepository := tasks_postgres_repository.NewTasksRepository(
		database.Pool,
	)
	tasksService := tasks_service.NewTaskService(tasksRepository)
	tasksHandler := tasks_transport_http.NewTasksHTTPHandler(tasksService)

	apiRouter := core_http_server.NewAPIVersionRouter(core_http_server.APIVersion1)
	apiRouter.RegisterRoutes(tasksHandler.Routes()...)

	rootMux := http.NewServeMux()
	rootMux.Handle(
		"/api/v1/",
		http.StripPrefix("/api/v1", apiRouter.WithMiddleware()),
	)

	logger := &core_logger.Logger{
		Logger: zap.NewNop(),
	}

	rootHandler := core_http_middleware.ChainMiddleware(
		rootMux,
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Panic(),
	)

	server := httptest.NewServer(rootHandler)
	t.Cleanup(server.Close)

	return server
}

func doJSONRequest(
	t *testing.T,
	client *http.Client,
	method string,
	url string,
	payload string,
) testHTTPResponse {
	t.Helper()

	requestContext, cancelRequest := context.WithTimeout(
		t.Context(),
		e2eRequestTimeout,
	)
	defer cancelRequest()

	var requestBody io.Reader
	if payload != "" {
		requestBody = strings.NewReader(payload)
	}

	request, err := http.NewRequestWithContext(
		requestContext,
		method,
		url,
		requestBody,
	)
	if err != nil {
		t.Fatalf("create %s request: %v", method, err)
	}

	if payload != "" {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("perform %s request: %v", method, err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read %s response body: %v", method, err)
	}

	return testHTTPResponse{
		StatusCode: response.StatusCode,
		Body:       responseBody,
	}
}

func assertHTTPStatus(
	t *testing.T,
	response testHTTPResponse,
	expectedStatus int,
) {
	t.Helper()

	if response.StatusCode != expectedStatus {
		t.Fatalf(
			"expected HTTP status %d, got %d; response body: %s",
			expectedStatus,
			response.StatusCode,
			response.Body,
		)
	}
}

func decodeJSON[T any](t *testing.T, body []byte) T {
	t.Helper()

	var result T

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("decode JSON response %q: %v", body, err)
	}

	return result
}
