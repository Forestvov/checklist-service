//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	tasks_transport_http "github.com/Forestvov/checklist-service/internal/features/tasks/transport/http"
	postgres_testutil "github.com/Forestvov/checklist-service/internal/testutil/postgres"
)

func TestTasksOverdueFilterE2E(t *testing.T) {
	setupContext, cancelSetup := context.WithTimeout(t.Context(), e2eSetupTimeout)
	database, err := postgres_testutil.Start(setupContext)
	cancelSetup()
	if err != nil {
		t.Fatalf("start test PostgreSQL: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("close test PostgreSQL: %v", err)
		}
	})

	server := newTasksTestServer(t, database)
	client := server.Client()
	now := time.Now().UTC()
	pastDueAt := now.Add(-24 * time.Hour)
	futureDueAt := now.Add(24 * time.Hour)

	overdueTask := createE2ETask(t, client, server.URL, "Overdue open task", &pastDueAt)
	futureTask := createE2ETask(t, client, server.URL, "Future open task", &futureDueAt)
	taskWithoutDeadline := createE2ETask(t, client, server.URL, "Task without deadline", nil)
	completedTask := createE2ETask(t, client, server.URL, "Completed past task", &pastDueAt)

	completePayload := fmt.Sprintf(`{"version":%d,"done":true}`, completedTask.Version)
	completeResponse := doJSONRequest(
		t,
		client,
		http.MethodPatch,
		fmt.Sprintf("%s/api/v1/tasks/%d", server.URL, completedTask.ID),
		completePayload,
	)
	assertHTTPStatus(t, completeResponse, http.StatusOK)

	overdueResponse := doJSONRequest(
		t,
		client,
		http.MethodGet,
		server.URL+"/api/v1/tasks?overdue=true",
		"",
	)
	assertHTTPStatus(t, overdueResponse, http.StatusOK)
	overdueTasks := decodeJSON[tasks_transport_http.GetTasksResponse](t, overdueResponse.Body)
	assertE2ETaskIDs(t, overdueTasks.Data, overdueTask.ID)

	notOverdueResponse := doJSONRequest(
		t,
		client,
		http.MethodGet,
		server.URL+"/api/v1/tasks?overdue=false",
		"",
	)
	assertHTTPStatus(t, notOverdueResponse, http.StatusOK)
	notOverdueTasks := decodeJSON[tasks_transport_http.GetTasksResponse](t, notOverdueResponse.Body)
	assertE2ETaskIDs(
		t,
		notOverdueTasks.Data,
		futureTask.ID,
		taskWithoutDeadline.ID,
		completedTask.ID,
	)
}

func createE2ETask(
	t *testing.T,
	client *http.Client,
	serverURL string,
	title string,
	dueAt *time.Time,
) tasks_transport_http.CreateTaskResponse {
	t.Helper()

	payload := map[string]any{"title": title}
	if dueAt != nil {
		payload["due_at"] = dueAt.Format(time.RFC3339Nano)
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("encode task payload: %v", err)
	}

	response := doJSONRequest(
		t,
		client,
		http.MethodPost,
		serverURL+"/api/v1/tasks",
		string(payloadBytes),
	)
	assertHTTPStatus(t, response, http.StatusCreated)

	return decodeJSON[tasks_transport_http.CreateTaskResponse](t, response.Body)
}

func assertE2ETaskIDs(
	t *testing.T,
	tasks []tasks_transport_http.TaskDTOResponse,
	expectedIDs ...int64,
) {
	t.Helper()

	if len(tasks) != len(expectedIDs) {
		t.Fatalf("expected %d tasks, got %d", len(expectedIDs), len(tasks))
	}

	actualIDs := make(map[int64]struct{}, len(tasks))
	for _, task := range tasks {
		actualIDs[task.ID] = struct{}{}
	}
	for _, expectedID := range expectedIDs {
		if _, exists := actualIDs[expectedID]; !exists {
			t.Errorf("expected task ID %d, got tasks %+v", expectedID, tasks)
		}
	}
}
