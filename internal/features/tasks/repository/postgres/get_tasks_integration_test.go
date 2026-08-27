//go:build integration

package tasks_postgres_repository

import (
	"testing"

	core_domain "github.com/Forestvov/checklist-service/internal/core/domain"
	core_pagination "github.com/Forestvov/checklist-service/internal/core/pagination"
)

func TestTasksRepositoryGetTasksEmpty(t *testing.T) {
	repository := newTestRepository(t)
	ctx := newTestContext(t)

	params, err := core_pagination.NewParams(nil, nil)
	if err != nil {
		t.Fatalf("create default pagination params: %v", err)
	}

	result, err := repository.GetTasks(ctx, params)
	if err != nil {
		t.Fatalf("get tasks from migrated database: %v", err)
	}
	if result.Total != 0 {
		t.Fatalf("unexpected total tasks: got %d, want 0", result.Total)
	}
	if result.Items == nil {
		t.Fatal("unexpected nil items, want empty slice")
	}
	if len(result.Items) != 0 {
		t.Fatalf("unexpected task count: got %d, want 0", len(result.Items))
	}
}

func TestTasksRepositoryGetTasksSuccess(t *testing.T) {
	repository := newTestRepository(t)
	ctx := newTestContext(t)

	titles := []string{
		"First task",
		"Second task",
		"Third task",
	}

	createdTasks := make([]core_domain.Task, 0, len(titles))

	for _, title := range titles {
		input := core_domain.NewTaskUninitialized(title, nil)

		created, err := repository.CreateTask(ctx, input)
		if err != nil {
			t.Fatalf("prepare task %q: %v", title, err)
		}

		createdTasks = append(createdTasks, created)
	}

	params, err := core_pagination.NewParams(nil, nil)
	if err != nil {
		t.Fatalf("create default pagination params: %v", err)
	}

	result, err := repository.GetTasks(ctx, params)
	if err != nil {
		t.Fatalf("get tasks: %v", err)
	}

	if result.Total != int64(len(createdTasks)) {
		t.Errorf(
			"unexpected total: got %d, want %d",
			result.Total,
			len(createdTasks),
		)
	}

	if len(result.Items) != len(createdTasks) {
		t.Fatalf(
			"unexpected task count: got %d, want %d",
			len(result.Items),
			len(createdTasks),
		)
	}

	for i, actual := range result.Items {
		expected := createdTasks[len(createdTasks)-1-i]

		if actual.ID != expected.ID {
			t.Errorf(
				"item %d: unexpected ID: got %d, want %d",
				i,
				actual.ID,
				expected.ID,
			)
		}

		if actual.Title != expected.Title {
			t.Errorf(
				"item %d: unexpected title: got %q, want %q",
				i,
				actual.Title,
				expected.Title,
			)
		}
	}

	if result.Params != params {
		t.Errorf(
			"unexpected params: got %+v, want %+v",
			result.Params,
			params,
		)
	}

	if result.TotalPages() != 1 {
		t.Errorf(
			"unexpected total pages: got %d, want 1",
			result.TotalPages(),
		)
	}
}

func TestTasksRepositoryGetTasksPagination(t *testing.T) {
	repository := newTestRepository(t)
	ctx := newTestContext(t)

	titles := []string{
		"First task",
		"Second task",
		"Third task",
		"Fourth task",
		"Fifth task",
	}

	createdTasks := make([]core_domain.Task, 0, len(titles))

	for _, title := range titles {
		input := core_domain.NewTaskUninitialized(title, nil)

		created, err := repository.CreateTask(ctx, input)
		if err != nil {
			t.Fatalf("prepare task %q: %v", title, err)
		}

		createdTasks = append(createdTasks, created)
	}

	page := 2
	perPage := 2

	params, err := core_pagination.NewParams(&page, &perPage)
	if err != nil {
		t.Fatalf("create pagination params: %v", err)
	}

	result, err := repository.GetTasks(ctx, params)
	if err != nil {
		t.Fatalf("get tasks: %v", err)
	}

	expectedTotal := int64(len(createdTasks))
	if result.Total != expectedTotal {
		t.Errorf(
			"unexpected total: got %d, want %d",
			result.Total,
			expectedTotal,
		)
	}

	if len(result.Items) != perPage {
		t.Fatalf(
			"unexpected task count: got %d, want %d",
			len(result.Items),
			perPage,
		)
	}

	const expectedTotalPages int64 = 3
	if result.TotalPages() != expectedTotalPages {
		t.Errorf(
			"unexpected total pages: got %d, want %d",
			result.TotalPages(),
			expectedTotalPages,
		)
	}

	if result.Params != params {
		t.Errorf(
			"unexpected params: got %+v, want %+v",
			result.Params,
			params,
		)
	}

	expectedTasks := []core_domain.Task{
		createdTasks[2],
		createdTasks[1],
	}

	for i, actual := range result.Items {
		expected := expectedTasks[i]

		if actual.ID != expected.ID {
			t.Errorf(
				"item %d: unexpected ID: got %d, want %d",
				i,
				actual.ID,
				expected.ID,
			)
		}

		if actual.Title != expected.Title {
			t.Errorf(
				"item %d: unexpected title: got %q, want %q",
				i,
				actual.Title,
				expected.Title,
			)
		}
	}
}
